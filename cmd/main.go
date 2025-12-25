package main

import (
	"fmt"

	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/spf13/pflag"
)

func main() {
	filepath := pflag.StringP("config", "c", ".app.config.dev.yaml", "configuration filepath (default: None)")
	migrate := pflag.Bool("migrate", false, "Update db structure")
	pflag.Parse()
	app.Load(*filepath)

	// Application entry point
	fmt.Printf("dsfgjhdg \n")
	app.Http.Server.Version = app.Version
}
