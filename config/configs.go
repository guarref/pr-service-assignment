package config

import (
	"fmt"
	"log"

	env "github.com/caarlos0/env/v6"
)

type Config struct {
	Port int `env:"PORT"`

	DBHost     string `env:"DB_HOST"`
	DBPort     string `env:"DB_PORT"`
	DBUser     string `env:"DB_USER"`
	DBPassword string `env:"DB_PASSWORD"`
	DBName     string `env:"DB_NAME"`

	MigrateEnable bool   `env:"MIGRATE_ENABLE"`
	MigrateFolder string `env:"MIGRATE_FOLDER"`
}

func Load() (*Config, error) {

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	return cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	return cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}
