package config

// config

type AppCfg struct {
	Logger   Logger         `json:"logger" yaml:"logger" env-prefix:"LOGGER_"`
	HTTP     HTTP           `json:"http" yaml:"http" env-prefix:"HTTP_"`
	TLS      TLS            `json:"tls" yaml:"tls" env-prefix:"TLS_"`
	Server   ServerConfig   `json:"server" yaml:"server" env-prefix:"SERVER_"`
	View     ViewConfig     `yaml:"view" `
	Database DatabaseConfig `yaml:"database"`
	Hash     Hash
}
