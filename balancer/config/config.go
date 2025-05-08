package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPConfig struct {
	Address string        `yaml:"address" env:"HTTP_ADDRESS" env-default:":8080"`
	Timeout time.Duration `yaml:"timeout" env:"API_TIMEOUT" env-default:"5s"`
}

type Config struct {
	LogLevel            string        `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	ServersURLs         string        `yaml:"servers_urls" env:"SERVERS_URLS" env-default:"http://localhost:8081,http://localhost:8082"`
	HealthCheckInterval time.Duration `yaml:"healthcheck_interval" env:"HEALTHCHECK_INTERVAL" env-default:"120s"`
	HTTPConfig          HTTPConfig    `yaml:"http"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
