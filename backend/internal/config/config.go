// config.go - Configuration management
// Handles environment variables, database connections, and app settings
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Provider ProviderConfig
	Search   SearchConfig
	Rate     RateLimitConfig
	Redis    RedisConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// ProviderConfig holds provider API URLs
type ProviderConfig struct {
	Provider1URL string
	Provider2URL string
}

// SearchConfig holds search-related configuration
type SearchConfig struct {
	MinFullTextLength         int
	CacheTTLSeconds           int
	QueryTimeoutSeconds       int // Timeout for search queries (default: 15)
	SimpleQueryTimeoutSeconds int // Timeout for simple queries like GetByID (default: 5)
}

// RateLimitConfig holds global rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
}

// RedisConfig holds Redis cache configuration
type RedisConfig struct {
	Enabled  bool
	Addr     string
	Password string
	DB       int
}

// Load reads environment variables and returns a Config struct
// This centralizes all configuration in one place, making it easy to manage
func Load() *Config {
	// Load .env file if it exists (optional, won't fail if missing)
	// This allows running with environment variables set directly
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "search_engine"),
		},
		Provider: ProviderConfig{
			Provider1URL: getEnv("PROVIDER1_URL", "https://raw.githubusercontent.com/WEG-Technology/mock/refs/heads/main/v2/provider1"),
			Provider2URL: getEnv("PROVIDER2_URL", "https://raw.githubusercontent.com/WEG-Technology/mock/refs/heads/main/v2/provider2"),
		},
		Search: SearchConfig{
			MinFullTextLength:         getEnvInt("SEARCH_MIN_FULLTEXT_LENGTH", 3),
			CacheTTLSeconds:           getEnvInt("SEARCH_CACHE_TTL_SECONDS", 60),
			QueryTimeoutSeconds:       getEnvInt("SEARCH_QUERY_TIMEOUT_SECONDS", 30),        // Increased to 30s for large datasets
			SimpleQueryTimeoutSeconds: getEnvInt("SEARCH_SIMPLE_QUERY_TIMEOUT_SECONDS", 10), // Increased to 10s
		},
		Rate: RateLimitConfig{
			RequestsPerMinute: getEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
		},
		Redis: RedisConfig{
			Enabled:  getEnvBool("REDIS_ENABLED", true),
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
	}
}

// getEnv retrieves an environment variable or returns a default value
// This provides a safe way to access environment variables with fallbacks
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt retrieves an environment variable as int or returns a default value
// If the value cannot be parsed, it falls back to the default
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return defaultValue
}

// getEnvBool retrieves an environment variable as bool or returns a default value.
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	switch value {
	case "1", "true", "TRUE", "True", "yes", "YES", "Yes":
		return true
	case "0", "false", "FALSE", "False", "no", "NO", "No":
		return false
	default:
		return defaultValue
	}
}

// GetDSN returns the MySQL Data Source Name string
// This formats the database connection string in MySQL format
func (c *Config) GetDSN() string {
	return c.Database.User + ":" + c.Database.Password + "@tcp(" + c.Database.Host + ":" + c.Database.Port + ")/" + c.Database.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
}

// Validate checks if required configuration values are present
// This helps catch configuration errors early
func (c *Config) Validate() error {
	// For now, we'll keep it simple and just log
	// In production, you might want to return errors for missing critical values
	log.Println("Configuration loaded successfully")
	return nil
}
