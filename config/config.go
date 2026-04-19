package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	AppAddr      string
	RedisAddr    string
	GeminiAPIKey string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR environment variable is required")
	}

	appAddr := os.Getenv("APP_ADDR")
	if appAddr == "" {
		log.Fatal("APP_ADDR environment variable is required")
	}

	return &Config{
		AppAddr:      appAddr,
		RedisAddr:    redisAddr,
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
	}
}
