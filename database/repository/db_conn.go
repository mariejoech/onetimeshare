package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

// ConnectDB opens a connection to the SQLite database and creates the table if it does not exist
func ConnectDB() error {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}

	// Get the database file path from environment variable
	dbFile := os.Getenv("DB_FILE")
	if dbFile == "" {
		return fmt.Errorf("DB_FILE environment variable not set")
	}

	// Open a connection to the SQLite database
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// Test the database connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("error pinging database: %v", err)
	}

	// Assign the database connection to the global variable
	DB = db
	fmt.Println("😁 Connected to SQLite database")

	// Ensure the table exists
	if err := ensureTableExists(); err != nil {
		return fmt.Errorf("error ensuring table exists: %v", err)
	}

	return nil
}

// ensureTableExists checks for the existence of the table and creates it if necessary
func ensureTableExists() error {
	// Define the table creation SQL query
	tableCreationQuery := `
	CREATE TABLE IF NOT EXISTS secrets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		token TEXT NOT NULL,
		text TEXT,
		password TEXT,
		expiration_date TIMESTAMP NOT NULL,
		is_burned BOOLEAN NOT NULL DEFAULT FALSE,
		is_viewed BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	// Execute the table creation query
	_, err := DB.Exec(tableCreationQuery)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	fmt.Println("🗂️ Table 'secrets' ensured.")
	return nil
}

// CloseDB closes the database connection when the application shuts down
func CloseDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Fatalf("error closing database connection: %v", err)
		}
	}
}
