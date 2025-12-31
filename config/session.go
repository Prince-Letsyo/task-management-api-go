package config

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis/v3"
)

type SessionConfig struct {
	Name           string        `yaml:"name" env:"SESSION_NAME" env-default:"fiber_session"`
	Lifetime       time.Duration `yaml:"lifetime" env:"SESSION_LIFETIME" env-default:"24h"`
	CookiePath     string        `yaml:"cookie_path" env:"SESSION_COOKIE_PATH" env-default:"/"`
	CookieDomain   string        `yaml:"cookie_domain" env:"SESSION_COOKIE_DOMAIN"`
	CookieSecure   bool          `yaml:"cookie_secure" env:"SESSION_COOKIE_SECURE" env-default:"false"`
	CookieHTTPOnly bool          `yaml:"cookie_http_only" env:"SESSION_COOKIE_HTTP_ONLY" env-default:"true"`
	CookieSameSite string        `yaml:"cookie_same_site" env:"SESSION_COOKIE_SAME_SITE" env-default:"Lax"`
	*session.Store
	// Redis backend settings
	Driver   string `yaml:"driver" env:"SESSION_DRIVER" env-required:"true"` // e.g., "redis"
	Host     string `yaml:"host" env:"SESSION_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"SESSION_PORT" env-default:"6379"`
	Password string `yaml:"password" env:"SESSION_PASSWORD"` // optional
	Username string `yaml:"username" env:"SESSION_USERNAME"` // for Redis 6+ ACL
	DB       int    `yaml:"db" env:"SESSION_DB" env-default:"0"`
}

func (s *SessionConfig) Setup() error {
	provider := redis.New(redis.Config{
		Host:     s.Host,
		Port:     s.Port,
		Username: s.Username,
		Password: s.Password,
		Database: s.DB,
	})

	s.Store = session.New(session.Config{
		Storage:        provider,
		KeyLookup:      "cookie:" + s.Name,
		CookiePath:     s.CookiePath,
		CookieDomain:   s.CookieDomain,
		CookieSecure:   s.CookieSecure,
		CookieHTTPOnly: s.CookieHTTPOnly,
		CookieSameSite: s.CookieSameSite,
		Expiration:     s.Lifetime,
	})
	return nil
}
