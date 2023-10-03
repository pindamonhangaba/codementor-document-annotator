package docs

import (
	"os"

	"github.com/pindamonhangaba/commandments"
	"github.com/pindamonhangaba/monoboi/backend/api/httpserver"

	"github.com/spf13/cobra"

	"github.com/labstack/echo/v4"
)

type Config struct {
	OutputFileApi string
}

func GenerateOpenAPI3() *cobra.Command {
	cmd := commandments.MustCMD("openapi", commandments.WithConfig(
		func(config Config) error {
			return genererateOpenAPISpec(config)
		}), commandments.WithDefaultConfig(Config{
		OutputFileApi: "./api.json",
	}))
	return cmd
}

func genererateOpenAPISpec(conf Config) error {
	e := echo.New()

	// register api routes
	servAPI := (&httpserver.HTTPServer{
		JWTConfig: httpserver.JWTConfig{
			Secret:       "secret",
			ClaimsCtxKey: "user",
		},
		ServerConf: httpserver.ServerConf{
			BodyLimit:     "50MB",
			Address:       "",
			AppAddress:    "",
			VersionString: "API 1.0",
		},
	}).Stub()
	oapi, err := servAPI.Register(e)
	if err != nil {
		return (err)
	}

	swagapijson, err := oapi.T().MarshalJSON()
	if err != nil {
		return err
	}
	err = os.WriteFile(conf.OutputFileApi, swagapijson, 0644)
	if err != nil {
		return err
	}

	return nil
}
