package config

import "os"

type Config struct {
	Addr        string
	DatabaseURL string
}

func Load() Config {
	return Config{
		Addr:        getenv("ADDR", ":8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://localhost:5432/webhook_gateway"),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
