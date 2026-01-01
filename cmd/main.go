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

	app.HTTP.Server.Version = app.Version
	if *migrate {
		model.Migrate(app.HTTP)
	} else {
		// app.LoadAdditionalServices() // Enable for PayPal and any other services
		core.LoadRoutes(app.HTTP.Server.App)
		app.HTTP.Route404()
		log.Fatal(app.HTTP.Server.ServeWithGraceFullShutdown(app.HTTP))
	}
}
