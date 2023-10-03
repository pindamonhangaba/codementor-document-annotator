package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/pindamonhangaba/apiculi/endpoint"
	"github.com/pindamonhangaba/monoboi/backend/postgres"
	"github.com/pindamonhangaba/monoboi/backend/service"
	"github.com/stretchr/testify/assert"
)

func TestSignin(t *testing.T) {
	e := echo.New()

	signing := postgres.LoginForm{
		Email:      "test@example.com",
		Password:   "123",
		ApplID:     "app",
		RememberMe: true,
	}
	signinJSON, err := json.Marshal(&signing)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/signin", strings.NewReader(string(signinJSON)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("_session_store", sessions.NewCookieStore([]byte("secret")))
	oapi := endpoint.NewOpenAPI("test", "v1")
	h := &AuthHandler{
		oapi: &oapi,
		signin: func(lf *postgres.LoginForm) (*service.AuthResponse, error) {
			assert.Equal(t, lf.ApplID, signing.ApplID)
			assert.Equal(t, lf.Password, signing.Password)
			assert.Equal(t, lf.RememberMe, signing.RememberMe)

			return &service.AuthResponse{
				User: service.User{
					ApplID: signing.ApplID,
				},
				Jwt: "ok",
			}, nil
		},
	}

	_, _, r := h.EmailLogin()
	if assert.NoError(t, r(c)) {
		if !assert.Equal(t, http.StatusOK, rec.Code) {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
		}
		c := rec.Result().Cookies()[0]
		if !assert.Equal(t, c.MaxAge, MaxAge) {
			t.Errorf("expected MaxAge %d, got %d", MaxAge, c.MaxAge)
		}
		if !c.HttpOnly {
			t.Errorf("expected HttpOnly true, got %t", true)
		}
		if !assert.Equal(t, c.SameSite, http.SameSiteNoneMode) {
			t.Errorf("expected SameSite %d, got %d", http.SameSiteNoneMode, c.SameSite)
		}
		if len(rec.Result().Header.Get("Set-Cookie")) < 1 {
			t.Errorf("expected a cookie value, got %s", rec.Result().Header.Get("Set-Cookie"))
		}
	}
}
