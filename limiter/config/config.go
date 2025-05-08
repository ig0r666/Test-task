package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPConfig struct {
	Address string        `yaml:"address" env:"ADDRESS" env-default:"localhost:8080"`
	Timeout time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"5s"`
}

type RateLimit struct {
	Capacity       int           `yaml:"capacity" env:"CAPACITY" env-default:"100"`
	UpdateInterval time.Duration `yaml:"update_interval" env:"UPDATE_INTERVAL" env-default:"1s"`
}

type Config struct {
	LogLevel   string     `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	DBAddress  string     `yaml:"db_address" env:"DB_ADDRESS"`
	RateLimit  RateLimit  `yaml:"ratelimiter"`
	HTTPConfig HTTPConfig `yaml:"http"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
