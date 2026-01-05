package config

import (
	"os"
	"time"
)

type Config struct {
	Port        string
	JWTSecret   string
	JWTExpiry   time.Duration
	DBPath      string
	CORSOrigin  string
	Environment string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}

	jwtExpiryStr := os.Getenv("JWT_EXPIRY")
	if jwtExpiryStr == "" {
		jwtExpiryStr = "24h"
	}
	jwtExpiry, _ := time.ParseDuration(jwtExpiryStr)
	if jwtExpiry == 0 {
		jwtExpiry, _ = time.ParseDuration("24h") // Default 24h
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/app.db"
	}

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		// По умолчанию разрешаем все origins (можно установить конкретные через CORS_ORIGIN)
		corsOrigin = "*"
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	return &Config{
		Port:        port,
		JWTSecret:   jwtSecret,
		JWTExpiry:   jwtExpiry,
		DBPath:      dbPath,
		CORSOrigin:  corsOrigin,
		Environment: env,
	}
}
