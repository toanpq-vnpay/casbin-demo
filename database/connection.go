package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func InitializeDatabase() error {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}

	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbUser := getEnvOrDefault("DB_USER", "postgres")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := getEnvOrDefault("DB_NAME", "postgres")

	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	fmt.Println("Database initialized successfully")
	return nil
}
