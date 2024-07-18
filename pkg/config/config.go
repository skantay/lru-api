// Package config provides a configuration mechanism for the application with specific policies.
// The configuration is loaded from environment variables
// and can be overridden by command-line flags
// If non of the sources of config containes any variables, then default variables are applied
package config

import (
	"flag"

	"github.com/caarlos0/env"
)

// Dedicated config struct for lru-api
type Config struct {
	HTTPPort        string `env:"HTTP_PORT" envDefault:"8080"`
	CacheSize       uint   `env:"CACHE_SIZE" envDefault:"10"`
	DefaultCacheTTL int64  `env:"DEFAULT_CACHE_TTL" envDefault:"60"`
	LogLevel        string `env:"LOG_LEVEL" envDefault:"WARN"`
}

// LoadConfig loads the configuration from environment variables.
func LoadConfig() (*Config, error) {
	httpPort := flag.String("server-host-port", "", "HTTP port")
	cacheSize := flag.Uint("cache-size", 0, "Cache size")
	defaultCacheTTL := flag.Int64("default-cache-ttl", 0, "Default cache TTL")
	logLevel := flag.String("log-level", "", "Log level")

	flag.Parse()

	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	if *httpPort != "" {
		cfg.HTTPPort = *httpPort
	}
	if *cacheSize != 0 {
		cfg.CacheSize = *cacheSize
	}
	if *defaultCacheTTL != 0 {
		cfg.DefaultCacheTTL = *defaultCacheTTL
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}

	return &cfg, nil
}
