package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string `env:"ENV" env-required:"true"`
	Tarantool  Tarantool
	Mattermost Mattermost
}

type Tarantool struct {
	Host     string `env:"TARANTOOL_HOST"`
	Port     uint16 `env:"TARANTOOL_PORT" env-default:"3301"`
	User     string `env:"TARANTOOL_USER"`
	Password string `env:"TARANTOOL_PASSWORD"`
}

type Mattermost struct {
	Token  string `env:"MM_TOKEN" env-required:"true"`
	Server string `env:"MM_SERVER" env-requried:"true"`
	Team   string `env:"MM_TEAM" env-required:"true"`
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
