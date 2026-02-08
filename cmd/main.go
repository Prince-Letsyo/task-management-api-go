package main

import (
	"log"

	"github.com/spf13/pflag"

	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/internal/core"
	"github.com/Prince-Letsyo/task-management-api-go/internal/queue"

	_ "ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	// Application entry point
	filepath := pflag.StringP("config", "c", ".app.config.dev.yaml", "configuration filepath (default: None)")
	// migrate := pflag.Bool("migrate", false, "Update db structure")
	pflag.Parse()
	app.Load(*filepath)

	app.HTTP.Server.Version = app.Version

	// Initialize Machinery
	machineryServer, err := queue.NewMachineryServer(app.HTTP.RabbitMQ, app.HTTP)
	if err != nil {
		log.Fatalf("Failed to initialize Machinery: %v", err)
	}

	// Create Machinery Client
	qClient := queue.NewWorkerClient(machineryServer)

	// Start Machinery Worker in background
	go func() {
		if err := queue.StartWorker(machineryServer, "auth_worker"); err != nil {
			log.Printf("Machinery worker error: %v", err)
		}
	}()

	// app.LoadAdditionalServices() // Enable for PayPal and any other services
	core.LoadRoutes(app.HTTP, qClient)
	app.HTTP.Route404()
	log.Fatal(app.HTTP.Server.ServeWithGraceFullShutdown(app.HTTP))
}
