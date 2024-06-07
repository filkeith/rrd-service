package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config contains all configs for different services.
type Config struct {
	// Logger
	LogLevel string `env:"LOG_LEVEL" env-default:"info"`
	// Http server params.
	HttpPort int `env:"HTTP_PORT" env-default:"8080"`
	// Storage paras
	StorageCapacity  uint64 `env:"STORAGE_CAP" env-default:"1000"`
	StorageHost      string `env:"STORAGE_HOST" env-default:"localhost"`
	StoragePort      int    `env:"STORAGE_PORT" env-default:"3000"`
	StorageNamespace string `env:"STORAGE_NAMESPACE" env-default:"test"`
}

// NewConfig returns initialized app config.
func NewConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load config from env: %w", err)
	}
	return &cfg, nil
}
