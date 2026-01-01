// Package model jhd
package model

import (
	"log"

	"github.com/Prince-Letsyo/task-management-api-go/config"
)

func Migrate(appConfig *config.AppCfg) {
	log.Println("Initiating migration...")
	err := appConfig.Database.DB.Migrator().AutoMigrate()
	if err != nil {
		panic(err)
	}
	log.Println("Migration Completed...")
}
