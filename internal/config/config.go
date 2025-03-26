package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Tarantool Tarantool
}

type Tarantool struct {
	Host     string `env:"TARANTOOL_HOST"`
	Port     uint16 `env:"TARANTOOL_PORT"`
	User     string `env:"TARANTOOL_USER"`
	Password string `env:"TARANTOOL_PASSWORD"`
}

func MustLoad() *Config {
	const op = "config.MustLoad"

	if err := godotenv.Load(); err != nil {
		log.Fatalf("%s: failed to load .env config: %v", op, err)
	}

	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("%s: failed to read config from env vars: %v", op, err)
	}

	return &cfg
}
