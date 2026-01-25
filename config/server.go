package config

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/oarkflow/log"
	"github.com/pkg/errors"
)

type ServerConfig struct {
	*fiber.App
	TemplateEngine *html.Engine

	Host    string        `mapstructure:"APP_HOST" yaml:"host" env:"APP_HOST" env-default:"localhost"`
	Port    string        `mapstructure:"APP_PORT" yaml:"port" env:"APP_PORT" env-default:"8080"`
	Timeout time.Duration `json:"timeout" yaml:"timeout" env:"TIMEOUT" env-default:"4s"`
	Prefix  string        `json:"prefix" yaml:"prefix" env:"PREFIX" env-default:""`
	APIPath string        `json:"api_path" yaml:"apiPath" env:"API_PATH" env-default:"/api"`

	Redirect    string `json:"redirect" mapstructure:"REDIRECT" yaml:"redirect" env:"REDIRECT" env-default:"http://localhost:3000"`
	Name        string `json:"name" mapstructure:"APP_NAME" yaml:"name" env:"APP_NAME" env-default:"iSend.to"`
	Version     string `json:"version" mapstructure:"APP_VERSION" yaml:"version" env:"APP_VERSION" env-default:"dev"`
	Mode        string `json:"mode" mapstructure:"APP_MODE" yaml:"mode" env:"APP_MODE" env-default:"app"`
	Env         string `json:"env" mapstructure:"APP_ENV" yaml:"env" env:"APP_ENV" env-default:"dev"`
	Key         string `json:"key" mapstructure:"APP_KEY" yaml:"key" env:"APP_KEY" env-default:"1894cde6c936a294a478cff0a9227fd276d86df6573b51af5dc59c9064edf426"`
	URL         string `json:"url" mapstructure:"APP_URL" yaml:"url" env:"APP_URL" env-default:"http://localhost"`
	Path        string `json:"path" mapstructure:"APP_PATH" yaml:"path" env:"APP_PATH"`
	ProxyHeader string `json:"PROXY_HEADER" mapstructure:"PROXY_HEADER" yaml:"PROXY_HEADER" env:"PROXY_HEADER" env-default:"*"`
	AssetPath   string `json:"asset_path" mapstructure:"ASSET_PATH" yaml:"asset_path" env:"ASSET_PATH" env-default:"assets"`
	PublicPath  string `json:"public_path" mapstructure:"PUBLIC_PATH" yaml:"public_path" env:"PUBLIC_PATH" env-default:"public"`
	UploadPath  string `json:"upload_path" mapstructure:"UPLOAD_PATH" yaml:"upload_path" env:"UPLOAD_PATH" env-default:"uploads"`
	StoragePath string `json:"storage_path" mapstructure:"STORAGE_PATH" yaml:"storage_path" env:"STORAGE_PATH" env-default:"storage"`
	LogPath     string `json:"log_path" mapstructure:"LOG_PATH" yaml:"log_path" env:"LOG_PATH" env-default:"storage/logs"`
	ExecPath    bool   `json:"exec_path" mapstructure:"EXEC_PATH" yaml:"exec_path" env:"EXEC_PATH" env-default:"false"`
	Debug       bool   `json:"debug" mapstructure:"APP_DEBUG" yaml:"debug" env:"APP_DEBUG" env-default:"true"`
	UploadSize  int    `json:"upload_size" mapstructure:"UPLOAD_SIZE" yaml:"upload_size" env:"UPLOAD_SIZE" env-default:"400"`
}

func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%v", s.Host, s.Port)
}

func (s ServerConfig) FullAPIPath() string {
	return fmt.Sprintf("%s%s", s.Prefix, s.APIPath)
}

func (s *ServerConfig) LoadPath() {
	if s.URL == "" {
		s.URL = fmt.Sprintf("http://localhost:%s", s.Port)
	}
	path, _ := os.Getwd()
	if s.ExecPath {
		path = getPath()
	}
	s.Path = path
	s.UploadPath = MakeDir(filepath.Join(path, s.UploadPath))
	s.AssetPath = MakeDir(filepath.Join(path, s.AssetPath))
	s.StoragePath = MakeDir(filepath.Join(path, s.StoragePath))
	s.LogPath = MakeDir(filepath.Join(path, s.LogPath))
	s.UploadSize = s.UploadSize * 1024 * 1024
}

func MakeDir(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.MkdirAll(path, os.ModePerm)
	}
	return path
}

func getPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return exPath
}

func (s *ServerConfig) Setup() {
	s.App = fiber.New(fiber.Config{
		Views:                 s.TemplateEngine,
		Concurrency:           256 * 1024 * 1024,
		ServerHeader:          s.Name,
		BodyLimit:             s.UploadSize,
		ReduceMemoryUsage:     true,
		ErrorHandler:          CustomErrorHandler,
		DisableStartupMessage: true,
		ProxyHeader:           s.ProxyHeader,
	})
}

func CustomErrorHandler(c *fiber.Ctx, err error) error {
	// StatusCode defaults to 500
	code := fiber.StatusInternalServerError
	//nolint:misspell    // Retrieve the custom statuscode if it's an fiber.*Error

	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	if c.Is("json") {
		return c.Status(code).JSON(err)
	}

	return c.Status(code).Render(fmt.Sprintf("error/%d", code), fiber.Map{ //nolint:nolintlint,errcheck
		"error": err,
	}, "layouts/main")
}

func (s *ServerConfig) ServeWithGraceFullShutdown(cfg *AppCfg, addr ...string) error {
	a := s.Host + ":" + s.Port
	if len(addr) != 0 {
		a = addr[0]
	}
	log.Info().Msgf("Starting server on %s", a)
	// Listen from a different goroutine
	go func() {
		if cfg.TLS.Cert.Filepath != "" &&
			cfg.TLS.Key.Filepath != "" {
			if err := s.ListenTLS(
				cfg.HTTP.Addr(),
				cfg.TLS.Cert.Filepath,
				cfg.TLS.Key.Filepath,
			); err != nil {
				log.Fatal().Err(err)
			}
		} else {
			if err := s.Listen(a); err != nil {
				log.Fatal().Err(err)
			}
		}
	}()

	c := make(chan os.Signal, 1) // Create channel to signify a signal being sent
	signal.Notify(c,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGQUIT,
	) // When an interrupt is sent, notify the channel
	<-c // This blocks the main thread until an interrupt is received
	fmt.Println("I'm shutting down")
	if err := s.Shutdown(); err != nil {
		_ = cfg.Database.Close()
		return err
	} else {
		_ = cfg.Database.Close()
	}
	return nil
}
