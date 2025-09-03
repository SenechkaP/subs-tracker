package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string
	AppPort    string
}

func LoadConfig(envPath string) *Config {
	if envPath != "" {
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("warning: could not load env file %s: %v", envPath, err)
		}
	}

	cfg := &Config{
		DBUser:     getEnv("POSTGRES_USER", "postgres"),
		DBPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		DBName:     getEnv("POSTGRES_DB", "substracker_db"),
		DBHost:     getEnv("POSTGRES_HOST", "localhost"),
		DBPort:     getEnv("POSTGRES_PORT", "5432"),
		AppPort:    getEnv("APP_PORT", "8080"),
	}
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
