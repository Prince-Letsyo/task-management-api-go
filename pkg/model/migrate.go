package model

import (
	"log"

	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
)

func Migrate() {
	log.Println("Initiating migration...")
	err := app.Http.Database.DB.Migrator().AutoMigrate()
	if err != nil {
		panic(err)
	}
	log.Println("Migration Completed...")
}
