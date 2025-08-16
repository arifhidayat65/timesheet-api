package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		Port:  getenv("PORT", "8080"),
		DB_DSN: getenv("DB_DSN", ""),
		Env:   getenv("APP_ENV", "development"),
		TZ:    getenv("TZ", "Asia/Jakarta"),
	}
	if cfg.DB_DSN == "" {
		log.Println("warning: DB_DSN empty")
	}
	return cfg
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
