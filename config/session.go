package config

import (
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis/v3"
)

type SessionConfig struct {
	*session.Store
	Driver string `yaml:"driver" env:"SESSION_DRIVER"`
	Name   string `yaml:"name" env:"SESSION_NAME"`
	Host   string `yaml:"host" env:"SESSION_HOST"`
	Port   int    `yaml:"port" env:"SESSION_PORT"`
	DB     int    `yaml:"db" env:"SESSION_DB"`
}

func (s *SessionConfig) Setup() error {
	provider := redis.New(redis.Config{
		Host:     s.Host,
		Port:     s.Port,
		Database: s.DB,
	})

	s.Store = session.New(session.Config{
		Storage: provider,
	})
	return nil
}
