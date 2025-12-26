package main

import (
	"log"

	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/internal/core"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/model"
	"github.com/spf13/pflag"
)

func main() {
	// Application entry point
	filepath := pflag.StringP("config", "c", ".app.config.dev.yaml", "configuration filepath (default: None)")
	migrate := pflag.Bool("migrate", false, "Update db structure")
	pflag.Parse()
	app.Load(*filepath)

	app.Http.Server.Version = app.Version
	if *migrate {
		model.Migrate()
	} else {
		// app.LoadAdditionalServices() // Enable for PayPal and any other services
		core.LoadRoutes(app.Http.Server.App)
		app.Http.Route404()
		log.Fatal(app.Http.Server.ServeWithGraceFullShutdown(app.Http))
	}
}
