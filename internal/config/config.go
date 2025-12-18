package config

import (
	"os"
)

type Config struct {
	RedisAddr string
	DBDSN     string
}

func Load() *Config {
	return &Config{
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		DBDSN:     getEnv("DB_DSN", "postgres://user:password@localhost:5432/agentmesh?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
