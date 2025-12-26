package config

import (
	"encoding/json"
	"time"

	"github.com/gofiber/storage/redis/v3"
)

type CacheConfig struct {
	*redis.Storage
	Driver string `yaml:"driver" env:"CACHE_DRIVER"`
	Name   string `yaml:"name" env:"CACHE_NAME"`
	Host   string `yaml:"host" env:"CACHE_HOST"`
	Port   int    `yaml:"port" env:"CACHE_PORT"`
	DB     int    `yaml:"db" env:"CACHE_DB"`
}

func (c *CacheConfig) Setup() {
	switch c.Driver {
	case "memcache":
	default:
		// Initialize custom config
		store := redis.New(redis.Config{
			Host:     c.Host,
			Port:     c.Port,
			Database: c.DB,
		})
		c.Storage = store
	}
}

func (c *CacheConfig) GetFromCache(tag string, item interface{}) error {
	if raw_data, err := c.Get(tag); err != nil {
		return err
	} else {
		if err := json.Unmarshal(raw_data, &item); err != nil {
			return err
		}
	}
	return nil
}

func (c *CacheConfig) SetToCache(name string, b interface{}) error {
	data, errB := json.Marshal(b)
	if errB != nil {
		return errB
	}

	if err := c.Set(name, data, time.Second*time.Duration(60)); err != nil {
		return err
	}
	return nil
}
