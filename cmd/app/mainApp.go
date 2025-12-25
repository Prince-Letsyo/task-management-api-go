package app

import "github.com/Prince-Letsyo/task-management-api-go/config"

var (
	Http    *config.AppCfg
	Version = "develop"
)

func Load(filePath string) {
}

func LoadBuiltInMiddlewares() {
}
