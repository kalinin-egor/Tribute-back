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

// ResetDatabase completely resets the database by dropping all tables and recreating them
func ResetDatabase() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	log.Println("🗑️ Resetting database...")

	// Drop all tables in correct order (due to foreign key constraints)
	tables := []string{
		"payments",
		"subscriptions",
		"channels",
		"users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("error dropping table %s: %w", table, err)
		}
		log.Printf("✅ Dropped table: %s", table)
	}

	// Create tables in correct order
	log.Println("📋 Creating tables...")

	// Create users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		user_id BIGINT PRIMARY KEY,
		earned NUMERIC(10, 2) DEFAULT 0.00,
		is_verified BOOLEAN DEFAULT FALSE,
		is_sub_published BOOLEAN DEFAULT FALSE,
		is_onboarded BOOLEAN DEFAULT FALSE
	);`

	if _, err := db.Exec(createUsersTable); err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}
	log.Println("✅ Created users table")

	// Create channels table
	createChannelsTable := `
	CREATE TABLE IF NOT EXISTS channels (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
		channel_username VARCHAR(255) UNIQUE NOT NULL
	);`

	if _, err := db.Exec(createChannelsTable); err != nil {
		return fmt.Errorf("error creating channels table: %w", err)
	}
	log.Println("✅ Created channels table")

	// Create subscriptions table
	createSubscriptionsTable := `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
		user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
		channel_username VARCHAR(255) NOT NULL,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		button_text VARCHAR(255),
		price NUMERIC(10, 2) NOT NULL,
		created_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createSubscriptionsTable); err != nil {
		return fmt.Errorf("error creating subscriptions table: %w", err)
	}
	log.Println("✅ Created subscriptions table")

	// Create payments table
	createPaymentsTable := `
	CREATE TABLE IF NOT EXISTS payments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
		description TEXT,
		created_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createPaymentsTable); err != nil {
		return fmt.Errorf("error creating payments table: %w", err)
	}
	log.Println("✅ Created payments table")

	log.Println("🎉 Database reset completed successfully!")
	return nil
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
