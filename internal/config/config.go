package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	EnvProduction  = "production"
	EnvDevelopment = "development"
)

type Config struct {
	AppEnv           string
	DbUrl            string
	AppPort          string
	JWTSecret        string
	TelegramBotToken string
	EnableCORS       bool
	AllowedOrigins   []string
	AppTLS           bool
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func (c *Config) checkRequiredFields() error {
	if c.TelegramBotToken == "" {
		return errors.New("TelegramBotToken environment variable not set")
	}
	if c.JWTSecret == "" {
		return errors.New("JWTSecret environment variable not set")
	}
	if c.DbUrl == "" {
		return errors.New("DB_URL environment variable not set")
	}
	return nil
}

func (c *Config) validate() error {
	// Common environment variables
	c.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	c.AppTLS = os.Getenv("APP_TLS") == "true"
	c.EnableCORS = os.Getenv("ENABLE_CORS") == "true"

	if c.EnableCORS {
		originsStr := os.Getenv("ALLOWED_ORIGINS")
		if originsStr == "" {
			return errors.New("CORS_ALLOWED_ORIGINS environment variable not set")
		}
		c.AllowedOrigins = strings.Split(originsStr, ",")
	}

	// Get strict environment variables
	switch c.AppEnv {
	case EnvDevelopment:
		c.DbUrl = os.Getenv("DEV_DB_URL")
		c.JWTSecret = os.Getenv("DEV_JWT_SECRET")
	case EnvProduction:
		c.DbUrl = os.Getenv("DB_URL")
		c.JWTSecret = os.Getenv("JWT_SECRET")
	default:
		return errors.New("unknown app environment")
	}

	if err := c.checkRequiredFields(); err != nil {
		return err
	}

	return nil
}

func Load() (*Config, error) {
	_ = godotenv.Load() //Don´t fail in case system variables such Docker or Kubernates

	env := getEnv("APP_ENV", EnvDevelopment)
	cfg := &Config{
		AppEnv:  env,
		AppPort: getEnv("APP_PORT", "8080"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
