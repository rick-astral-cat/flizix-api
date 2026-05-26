package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

const (
	EnvProduction  = "production"
	EnvDevelopment = "development"
)

type Config struct {
	AppEnv    string
	DbUrl     string
	AppPort   string
	JWTSecret string
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func (c *Config) validateDev() error {
	dbUrl := os.Getenv("DEV_DB_URL")
	if dbUrl == "" {
		return errors.New("DEV_DB_URL environment variable not set")
	}
	c.DbUrl = dbUrl
	c.AppEnv = "development"

	jwtSecret := os.Getenv("DEV_JWT_SECRET")
	if jwtSecret == "" {
		return errors.New("DEV_JWT_SECRET environment variable not set")
	}
	c.JWTSecret = jwtSecret

	return nil
}

func (c *Config) validateProd() error {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		return errors.New("DB_URL environment variable not set")
	}
	c.DbUrl = dbUrl

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return errors.New("JWT_SECRET environment variable not set")
	}
	c.JWTSecret = jwtSecret

	return nil
}

func Load() (*Config, error) {
	_ = godotenv.Load() //Don´t fail in case system variables such Docker or Kubernates

	cfg := &Config{
		AppEnv:  "",
		DbUrl:   "",
		AppPort: getEnv("APP_PORT", "8080"),
	}

	env := os.Getenv("APP_ENV")
	switch env {
	case EnvProduction:
		cfg.AppEnv = EnvProduction
		err := cfg.validateProd()
		if err != nil {
			return nil, err
		}
	case EnvDevelopment:
		cfg.AppEnv = EnvDevelopment
		err := cfg.validateDev()
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid APP_ENV")
	}

	return cfg, nil
}
