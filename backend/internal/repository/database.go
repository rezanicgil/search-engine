// database.go - Database connection management
// Handles MySQL connection, connection pooling, and provides a shared database instance
package repository

import (
	"database/sql"
	"fmt"
	"log"
	"search-engine/backend/internal/config"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver - imported for side effects
)

// DB is the global database connection instance
// This allows all repositories to share the same connection pool
var DB *sql.DB

// Connect initializes the database connection using configuration
// This sets up connection pooling and connection parameters for optimal performance
func Connect(cfg *config.Config) error {
	// Get the Data Source Name (DSN) from config
	// DSN format: user:password@tcp(host:port)/dbname?params
	dsn := cfg.GetDSN()

	// Open a database connection
	// sql.Open doesn't actually connect - it just prepares the connection
	// The actual connection happens on the first query
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool settings
	// These are important for performance and resource management

	// SetMaxOpenConns sets the maximum number of open connections to the database
	// Too high = resource exhaustion, too low = connection starvation
	DB.SetMaxOpenConns(25)

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
	// Keeping some idle connections ready improves response time
	DB.SetMaxIdleConns(5)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused
	// This prevents using stale connections that might have been closed by the server
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection by pinging the database
	// This ensures the connection string is correct and database is accessible
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")
	return nil
}

// Close closes the database connection
// Should be called during application shutdown to clean up resources
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB returns the global database instance
// This allows other packages to access the database connection
// In a more complex app, you might use dependency injection instead
func GetDB() *sql.DB {
	return DB
}
