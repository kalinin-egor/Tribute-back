package database

import (
	"database/sql"
	"fmt"
	"log"

	"tribute-back/internal/config"

	_ "github.com/lib/pq"
)

var db *sql.DB

// Init initializes the database connection
func Init() (*sql.DB, error) {
	cfg := config.GetDatabaseConfig()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Database connected successfully")
	return db, nil
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return db
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
