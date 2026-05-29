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

	c.EnableCORS = os.Getenv("ENABLE_CORS") == "true"
	if c.EnableCORS {
		originsStr := os.Getenv("ALLOWED_ORIGINS")
		if originsStr == "" {
			return errors.New("CORS_ALLOWED_ORIGINS environment variable not set")
		}
		c.AllowedOrigins = strings.Split(originsStr, ",")
	}

	jwtSecret := os.Getenv("DEV_JWT_SECRET")
	if jwtSecret == "" {
		return errors.New("DEV_JWT_SECRET environment variable not set")
	}
	c.JWTSecret = jwtSecret

	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if tgToken == "" {
		return errors.New("TELEGRAM_BOT_TOKEN environment variable not set")
	}
	c.TelegramBotToken = tgToken

	return nil
}

func (c *Config) validateProd() error {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		return errors.New("DB_URL environment variable not set")
	}
	c.DbUrl = dbUrl

	c.EnableCORS = os.Getenv("ENABLE_CORS") == "true"
	if c.EnableCORS {
		originsStr := os.Getenv("ALLOWED_ORIGINS")
		if originsStr == "" {
			return errors.New("CORS_ALLOWED_ORIGINS environment variable not set")
		}
		c.AllowedOrigins = strings.Split(originsStr, ",")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return errors.New("JWT_SECRET environment variable not set")
	}
	c.JWTSecret = jwtSecret

	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if tgToken == "" {
		return errors.New("TELEGRAM_BOT_TOKEN environment variable not set")
	}
	c.TelegramBotToken = tgToken

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
