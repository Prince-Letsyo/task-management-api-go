package main

import (
	"log"

	"github.com/spf13/pflag"

	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/internal/core"

	_ "ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	// Application entry point
	filepath := pflag.StringP("config", "c", ".app.config.dev.yaml", "configuration filepath (default: None)")
	// migrate := pflag.Bool("migrate", false, "Update db structure")
	pflag.Parse()
	app.Load(*filepath)

	app.HTTP.Server.Version = app.Version

	// app.LoadAdditionalServices() // Enable for PayPal and any other services
	core.LoadRoutes(app.HTTP)
	app.HTTP.Route404()
	log.Fatal(app.HTTP.Server.ServeWithGraceFullShutdown(app.HTTP))
}
