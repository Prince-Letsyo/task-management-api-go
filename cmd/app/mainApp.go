// Package app dd
package app

import (
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/helmet/v2"
	"github.com/mikhail-bigun/fiberlogrus"

	"github.com/Prince-Letsyo/task-management-api-go/config"
)

var (
	HTTP    *config.AppCfg
	Version = "develop"
)

func Load(filepath string) {
	HTTP = &config.AppCfg{ConfigFile: filepath}
	HTTP.SetUp()
	LoadBuiltInMiddlewares(HTTP)
}

func LoadBuiltInMiddlewares(app *config.AppCfg) {
	app.Server.Use(recover.New())
	app.Server.Use(cors.New(
		cors.ConfigDefault,
	))
	app.Server.Use(helmet.New())
	app.Server.Use(requestid.New())
	app.Server.Use(etag.New())
	app.Server.Use(compress.New(compress.Config{
		Level: 1,
	}))
	if app.Server.Debug {
		app.Server.Use(pprof.New())
		app.Server.Use(fiberlogrus.New(fiberlogrus.Config{
			Logger: app.Log.NewLogger(),
			Tags: []string{
				fiberlogrus.TagLatency,
				fiberlogrus.TagMethod,
				fiberlogrus.TagURL,
				fiberlogrus.TagUA,
				fiberlogrus.TagBytesSent,
				fiberlogrus.TagPid,
				fiberlogrus.TagStatus,
			},
		}))

	}
}
