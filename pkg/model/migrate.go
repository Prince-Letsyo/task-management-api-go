// Package model jhd
package model

import (
	"log"

	"github.com/Prince-Letsyo/task-management-api-go/config"
)

func Migrate(appCfg *config.AppCfg) {
	log.Println("Initiating migration...")
	err := appCfg.Database.DB.Migrator().AutoMigrate()
	if err != nil {
		panic(err)
	}
	log.Println("Migration Completed...")
}
