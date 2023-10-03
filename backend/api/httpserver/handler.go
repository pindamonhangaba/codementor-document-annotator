package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/pindamonhangaba/apiculi/endpoint"
	"github.com/pindamonhangaba/monoboi/backend/api/httpserver/middleware"
	"github.com/pindamonhangaba/monoboi/backend/lib"
	"github.com/pindamonhangaba/monoboi/backend/mailmaid"
	"github.com/pindamonhangaba/monoboi/backend/postgres"
	"github.com/pindamonhangaba/monoboi/backend/service"
	"github.com/pkg/errors"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	mw "github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth/gothic"
)

// Based on Google JSONC styleguide
// https://google.github.io/styleguide/jsoncstyleguide.xml

type ContextQ struct {
	Context string `json:"context,omitempty"`
}

// JWTConfig holds configuration for JWT context
type JWTConfig struct {
	Secret,
	ClaimsCtxKey string
}

// ServerConf holds configuration data for the webserver
type ServerConf struct {
	BodyLimit            string
	Address              string
	AppAddress           string
	VersionString        string
	AllowedOrigins       []string
	CEPAbertoAPIToken    string
	PagarmeAPIKey        string
	OwnPublicURL         string
	OauthRedirectURL     string
	PostbackVerification string
}

// HTTPServer create a service to echo server
type HTTPServer struct {
	stub       bool
	DB         *sqlx.DB
	JWTConfig  JWTConfig
	ServerConf ServerConf
	Mailer     *mailmaid.Mailer
}

func (h *HTTPServer) Stub() HTTPServer {
	stub := *h
	stub.stub = true
	return stub
}

func (h *HTTPServer) validateDeps() error {
	if h.DB == nil || h.Mailer == nil {
		return errors.New("missing dependency")
	}
	return nil
}

// Run create a new echo server
func (h *HTTPServer) Register(e *echo.Echo) (*endpoint.OpenAPI, error) {
	err := h.validateDeps()
	if !h.stub && err != nil {
		return nil, err
	}

	// Echo instance
	e.Use(mw.Recover())
	e.Use(mw.Logger())
	e.Use(mw.BodyLimit("50M"))
	e.HTTPErrorHandler = lib.HTTPErrorHandler

	e.Use(mw.CORSWithConfig(mw.CORSConfig{
		AllowOrigins:     h.ServerConf.AllowedOrigins,
		AllowMethods:     []string{echo.GET, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowCredentials: true,
	}))
	ratedBase := e.Group("")
	ratedBase.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
		Limit: 10,
		Burst: 20,
	}))

	gJWT := e.Group("")
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(service.Claims)
		},
		SigningKey:    []byte(h.JWTConfig.Secret),
		ContextKey:    h.JWTConfig.ClaimsCtxKey,
		SigningMethod: jwt.SigningMethodHS256.Name,
	}
	jwtServiceConfig := postgres.JWTConfig{
		Secret:          h.JWTConfig.Secret,
		HoursTillExpire: time.Minute * 5,
		SigningMethod:   jwt.SigningMethodHS256,
	}

	gJWT.Use(echojwt.WithConfig(jwtConfig))
	ratedAPI := gJWT.Group("")
	ratedAPI.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
		Limit: 10,
		Burst: 20,
	}))
	ratedAPI.Use(echojwt.WithConfig(jwtConfig))

	gONB := e.Group("")
	jwtConfigOnb := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(service.Claims)
		},
		SigningKey: []byte(h.JWTConfig.Secret),
		ContextKey: h.JWTConfig.ClaimsCtxKey,
		Skipper: func(c echo.Context) bool {
			auth := c.Request().Header.Get("Authorization")
			return len(auth) <= 0
		},
		SigningMethod: jwt.SigningMethodHS256.Name,
	}
	gONB.Use(echojwt.WithConfig(jwtConfigOnb))

	oapi := endpoint.NewOpenAPI("API", "v1")
	oapi.AddServer(e.Server.Addr, "current server")

	gauth := e.Group("")
	gauth.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
		Limit: 2,
		Burst: 2,
	}))
	gauth.Use(session.Middleware(sessions.NewCookieStore([]byte(h.JWTConfig.Secret))))

	/*
	 * public routes
	 */

	pv := postgres.PwdRecoverer{
		DB:     h.DB,
		Mailer: h.Mailer,
		Config: struct{ AuthURL string }{AuthURL: h.ServerConf.AppAddress + "/api/auth"},
	}
	pr := postgres.PwdReseter{
		DB:     h.DB,
		Mailer: h.Mailer,
	}
	gpr := postgres.GetActionVerification{DB: h.DB}
	eta := postgres.EmailTokenAuth{
		DB:     h.DB,
		Mailer: h.Mailer,
	}

	ett := postgres.EmailTokenAuthenticator{
		DB:        h.DB,
		Mailer:    h.Mailer,
		JWTConfig: jwtServiceConfig,
	}

	a := &postgres.Authenticator{
		DB:        h.DB,
		Logger:    log.New(os.Stdout, "Auth: ", log.LstdFlags),
		JWTConfig: jwtServiceConfig,
		Mailer:    h.Mailer,
	}

	ah := &AuthHandler{
		oapi:                  &oapi,
		signin:                a.EmailLogin,
		refreshToken:          a.RefreshToken,
		oneTimeLogin:          a.OneTimeLogin,
		emailTokenSignin:      ett.Run,
		getActionVerification: gpr.Run,
		emailTokenAuth:        eta.Run,
		pwdReset:              pr.Run,
		pwdRecover:            pv.Run,
	}
	gauth.Add(ah.EmailLogin())
	gauth.Add(ah.PasswordRecover())
	gauth.Add(ah.PasswordReset())
	gauth.Add(ah.GetActionVerification())
	gauth.Add(ah.OneTimeLogin())
	gauth.Add(ah.EmailTokenAuth())
	gauth.Add(ah.EmailTokenSignin())
	gauth.Add(ah.RefreshJWT())
	gauth.Add(ah.Logout())

	trdp := postgres.ThirdPartyOnboarder{
		DB:        h.DB,
		JWTConfig: jwtServiceConfig,
		Logger:    log.New(os.Stdout, "3rd Party Oauth: ", log.LstdFlags),
		Mailer:    h.Mailer,
	}

	tpoa := ThirdPartyOauthHandler{
		oapi:                  &oapi,
		completeUserAuth:      gothic.CompleteUserAuth,
		beginAuth:             gothic.BeginAuthHandler,
		onboardThirdPartyUser: trdp.Run,
		oauthRedirectURL:      h.ServerConf.OauthRedirectURL,
	}

	gauth.Add(tpoa.OauthProviderFlow())
	gauth.Add(tpoa.OauthProviderCallback())

	oc := &postgres.OnboardingCreator{
		DB: h.DB,
	}
	oh := &OnboardingHandler{
		oapi:                   &oapi,
		createUser:             oc.Run,
		authenticateEmailLogin: a.EmailLogin,
	}
	gauth.Add(oh.CreateAccount())

	ugs := postgres.UserGetter{DB: h.DB}
	ups := postgres.UserPatch{DB: h.DB}
	umgs := postgres.UserMeGetter{DB: h.DB}
	uh := UserHandler{
		oapi:          &oapi,
		getUserByID:   ugs.Run,
		patchUserByID: ups.Run,
		getUserMe:     umgs.Run,
	}

	ratedAPI.Add(uh.GetUserByID())
	ratedAPI.Add(uh.PatchUserByID())
	ratedAPI.Add(uh.GetUserMe())

	// Geo location routes
	geloS := &postgres.SearchGeoLocationsService{DB: h.DB}
	geloU := &postgres.UpdateZipCodeService{DB: h.DB}
	cpG := &lib.CepGetterService{
		CEPAbertoAPIToken: h.ServerConf.CEPAbertoAPIToken,
		UpdateZipCode:     geloU,
	}
	geloH := &GeoLocationHandler{
		oapi:               &oapi,
		searchGeoLocations: geloS.Run,
		zipcodeUpdate:      cpG.Run,
	}
	e.Add(geloH.SearchGeoLocations())

	swagapijson, err := oapi.T().MarshalJSON()
	if err != nil {
		panic(err)
	}

	e.GET("/docs/swagger.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, json.RawMessage(swagapijson))
	})
	e.GET("/docs", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Redoc</title>
				<!-- needed for adaptive design -->
				<meta charset="utf-8"/>
				<meta name="viewport" content="width=device-width, initial-scale=1">
				<link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
				<!--
				Redoc doesn't change outer page styles
				-->
				<style>
				body {
					margin: 0;
					padding: 0;
				}
				</style>
			</head>
			<body>
				<redoc spec-url='/docs/swagger.json'></redoc>

				<script src="https://cdn.jsdelivr.net/npm/redoc@latest/bundles/redoc.standalone.js"> </script>
			</body>
			</html>
		`)
	})
	return &oapi, nil
}
