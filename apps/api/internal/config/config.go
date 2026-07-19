package config

import "os"

type Config struct {
	AppName       string
	HTTPAddr      string
	DatabaseURL   string
	JWTSecret     string
	AdminEmail    string
	AdminPassword string
}

func Load() Config {
	return Config{
		AppName:       getEnv("APP_NAME", "fungi-wiki-api"),
		HTTPAddr:      getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://fungi:fungi@localhost:55432/fungi_wiki?sslmode=disable"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-only-change-this-secret"),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@fungi.local"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123456"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
