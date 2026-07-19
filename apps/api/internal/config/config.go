package config

import "os"

type Config struct {
	AppName     string
	HTTPAddr    string
	DatabaseURL string
}

func Load() Config {
	return Config{
		AppName:     getEnv("APP_NAME", "fungi-wiki-api"),
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://fungi:fungi@localhost:55432/fungi_wiki?sslmode=disable"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
