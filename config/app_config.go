// Package config ds
package config

// config
import (
	logger "log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/oarkflow/log"
	"github.com/pkg/errors"

	"github.com/Prince-Letsyo/task-management-api-go/pkg"
)

type AppCfg struct {
	Logger     Logger         `json:"logger" yaml:"logger" env-prefix:"LOGGER_"`
	HTTP       HTTP           `json:"http" yaml:"http" env-prefix:"HTTP_"`
	TLS        TLS            `json:"tls" yaml:"tls" env-prefix:"TLS_"`
	Server     ServerConfig   `json:"server" yaml:"server" env-prefix:"SERVER_"`
	View       ViewConfig     `yaml:"view" `
	JwtSecrets JwtSecrets     `yaml:"jwt"`
	Database   DatabaseConfig `yaml:"database"`
	Mail       Mail           `yaml:"mail"`
	Redis      Redis          `json:"redis" yaml:"redis" env-prefix:"REDIS_"`
	Postgres   Postgres       `json:"postgres" yaml:"postgres" env-prefix:"POSTGRES_"`
	Swagger    Swagger        `json:"swagger" yaml:"swagger" env-prefix:"SWAGGER_"`
	Log        LogConfig      `yaml:"log"`
	Cache      CacheConfig    `yaml:"cache"`
	Storage    StorageConfig  `yaml:"storage"`
	Session    SessionConfig  `yaml:"session"`
	Auth       AuthSettings   `json:"auth" yaml:"auth" env-prefix:"AUTH_"`
	ConfigFile string
	Encryptor  *pkg.Encryptor
}

func (cfg *AppCfg) Route404() {
	cfg.Server.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).SendString("Page not found")
	})
}

func (cfg *AppCfg) SetUp() {
	err := cfg.newAppCfg(cfg.ConfigFile)
	if err != nil {
		logger.Fatalf("cannot load config: %s", err)
	}
	cfg.Server.LoadPath()
	cfg.View.Load(cfg.Server.Path)
	cfg.Mail.View = &cfg.View
	cfg.Server.TemplateEngine = cfg.View.Template.TemplateEngine
	cfg.Server.Setup()
	cfg.LoadComponents()
}

func (cfg *AppCfg) newAppCfg(filepath string) error {
	var err error
	if filepath == "" {
		err = cleanenv.ReadEnv(cfg)
		if err != nil {
			return errors.Wrap(err, "cannot read env")
		}
	} else {
		err = cleanenv.ReadConfig(filepath, cfg)
		if err != nil {
			return errors.Wrap(err, "cannot read config")
		}
	}
	return err
}

func (cfg *AppCfg) LoadComponents() {
	cfg.LoadStatic()
	cfg.PrepareLog()
	_ = cfg.Database.Setup()
	_ = cfg.Session.Setup()
	cfg.Cache.Setup()
	cfg.Storage.Setup()

	var err error
	cfg.Encryptor, err = pkg.NewEncryptor(cfg.Server.Key)
	if err != nil {
		logger.Fatalf("cannot create encryptor: %s", err)
	}
}

func (cfg *AppCfg) LoadStatic() {
	cfg.Server.Static("/websocket", "./resources/views/websocket.html")
	cfg.Server.Static("/", filepath.Join(cfg.Server.Path, cfg.Server.PublicPath), fiber.Static{
		Compress:      true,
		ByteRange:     true,
		CacheDuration: 24 * time.Hour,
	})
}

func (cfg *AppCfg) PrepareLog() {
	writer := &log.MultiWriter{}
	path := MakeDir(filepath.Join(cfg.Server.Path, cfg.Log.InfoLevel.Path))
	writer.InfoWriter = &log.FileWriter{Filename: filepath.Join(path, "INFO.log"), EnsureFolder: true, TimeFormat: cfg.Log.InfoLevel.TimeFormat}

	_ = MakeDir(filepath.Join(cfg.Server.Path, cfg.Log.WarnLevel.Path))
	writer.WarnWriter = &log.FileWriter{Filename: filepath.Join(cfg.Server.Path, cfg.Log.WarnLevel.Path, "WARN.log"), EnsureFolder: true, TimeFormat: cfg.Log.WarnLevel.TimeFormat}

	_ = MakeDir(filepath.Join(cfg.Server.Path, cfg.Log.ErrorLevel.Path))
	writer.ErrorWriter = &log.FileWriter{Filename: filepath.Join(cfg.Server.Path, cfg.Log.ErrorLevel.Path, "ERROR.log"), EnsureFolder: true, TimeFormat: cfg.Log.ErrorLevel.TimeFormat}
	if cfg.Log.ConsoleLog.Show {
		writer.ConsoleWriter = &log.IOWriter{Writer: os.Stderr}
		writer.ConsoleLevel = log.InfoLevel
	}
	log.DefaultLogger = log.Logger{
		TimeField:  cfg.Log.TimeField,
		TimeFormat: cfg.Log.TimeFormat,
		Writer:     writer,
	}
}
