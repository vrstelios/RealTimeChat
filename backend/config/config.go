package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	AppAddr        string
	RedisAddr      string
	GeminiAPIKey   string
	MongoURI       string
	QdrantHost     string
	QdrantPort     int
	JaegerEndpoint string
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

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	mongoURL := os.Getenv("MONGO_URI")
	if mongoURL == "" {
		log.Fatal("MONGO_URI environment variable is required")
	}

	QdrantHost := os.Getenv("QDRANT_HOST")
	if QdrantHost == "" {
		log.Fatal("QDRANT_HOST environment variable is required")
	}

	QdrantPort := os.Getenv("QDRANT_PORT")
	if QdrantPort == "" {
		log.Fatal("QDRANT_PORT environment variable is required")
	}

	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		log.Fatal("JAEGER_ENDPOINT environment variable is required")
	}

	port, err := strconv.Atoi(QdrantPort)
	if err != nil {
		log.Fatal("Invalid QDRANT_PORT environment variable:", err)
	}

	return &Config{
		AppAddr:        appAddr,
		RedisAddr:      redisAddr,
		GeminiAPIKey:   geminiAPIKey,
		MongoURI:       mongoURL,
		QdrantHost:     QdrantHost,
		QdrantPort:     port,
		JaegerEndpoint: jaegerEndpoint,
	}
}
