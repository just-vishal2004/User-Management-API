package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the application.
// Every other package receives this struct — nothing reads
// environment variables directly except this package.
type Config struct {
	AppPort string
	AppEnv  string
	DBUrl   string
}

// Load reads the .env file (if present), then reads environment
// variables and returns a populated Config struct.
// It is called once at application startup in main.go.
func Load() *Config {
	// Load .env file if it exists.
	// In production, environment variables are set directly
	// in the shell or container — no .env file is present,
	// so we don't fatal if the file is missing.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	cfg := &Config{
		AppPort: getEnv("APP_PORT", "3000"),
		AppEnv:  getEnv("APP_ENV", "development"),
		DBUrl:   buildDBUrl(),
	}

	return cfg
}

// buildDBUrl constructs the PostgreSQL connection string
// from individual environment variables.
// Format: postgres://user:password@host:port/dbname?sslmode=disable
func buildDBUrl() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "ainyx_users")
	sslmode := getEnv("DB_SSLMODE", "disable")

	// If no password is set, omit it from the connection string.
	// pgx handles an empty password differently from no password.
	if password == "" {
		return fmt.Sprintf(
			"postgres://%s@%s:%s/%s?sslmode=%s",
			user, host, port, dbname, sslmode,
		)
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode,
	)
}

// getEnv reads an environment variable by key.
// If the variable is not set, it returns the provided fallback value.
// This prevents the app from panicking on missing optional config.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
