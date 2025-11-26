// migrator.go - Database migration system
// Handles running SQL migration files and tracking which migrations have been applied
package migration

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migrator handles database migrations
// This struct encapsulates migration logic and state
type Migrator struct {
	db            *sql.DB
	migrationsDir string
}

// NewMigrator creates a new migration runner
// db: database connection
// migrationsDir: path to directory containing SQL migration files
func NewMigrator(db *sql.DB, migrationsDir string) *Migrator {
	return &Migrator{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

// Run executes all pending migrations
// This is the main entry point for running migrations
// It ensures migrations run in order and are only applied once
func (m *Migrator) Run() error {
	log.Println("Starting database migrations...")

	// Ensure migration_history table exists
	// This table tracks which migrations have been applied
	if err := m.ensureMigrationTable(); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// Get list of migration files
	// We read all .sql files from the migrations directory
	migrations, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get list of already applied migrations
	// This prevents running the same migration twice
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Filter out already applied migrations
	// Only run migrations that haven't been executed yet
	pending := m.filterPendingMigrations(migrations, applied)

	if len(pending) == 0 {
		log.Println("No pending migrations found")
		return nil
	}

	log.Printf("Found %d pending migration(s)", len(pending))

	// Execute each pending migration in order
	// Migrations are sorted by filename to ensure correct execution order
	for _, migration := range pending {
		if err := m.runMigration(migration); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration, err)
		}
		log.Printf("✓ Applied migration: %s", migration)
	}

	log.Println("All migrations completed successfully")
	return nil
}

// ensureMigrationTable creates the migration_history table if it doesn't exist
// This table is created manually in the first migration, but we check here for safety
func (m *Migrator) ensureMigrationTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migration_history (
			id INT AUTO_INCREMENT PRIMARY KEY,
			migration_name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_migration_name (migration_name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err := m.db.Exec(query)
	return err
}

// getMigrationFiles reads all .sql files from the migrations directory
// Returns sorted list of migration filenames
func (m *Migrator) getMigrationFiles() ([]string, error) {
	var files []string

	// Read directory
	entries, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		return nil, err
	}

	// Filter .sql files and extract names
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	// Sort files to ensure correct execution order
	// Files are named like 001_initial_schema.sql, 002_add_index.sql, etc.
	sort.Strings(files)

	return files, nil
}

// getAppliedMigrations queries the database for already applied migrations
// Returns a map for fast lookup
func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := m.db.Query("SELECT migration_name FROM migration_history")
	if err != nil {
		// If table doesn't exist yet, return empty map (first migration will create it)
		if strings.Contains(err.Error(), "doesn't exist") {
			return applied, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}

	return applied, rows.Err()
}

// filterPendingMigrations returns only migrations that haven't been applied yet
func (m *Migrator) filterPendingMigrations(all []string, applied map[string]bool) []string {
	var pending []string
	for _, migration := range all {
		if !applied[migration] {
			pending = append(pending, migration)
		}
	}
	return pending
}

// runMigration executes a single migration file
// Reads the SQL file, executes it, and records it in migration_history
func (m *Migrator) runMigration(filename string) error {
	// Read migration file
	path := filepath.Join(m.migrationsDir, filename)
	sqlContent, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Begin transaction
	// This ensures migration is atomic - either fully succeeds or fully rolls back
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	statements, err := parseSQLStatements(string(sqlContent))
	if err != nil {
		return fmt.Errorf("failed to parse migration %s: %w", filename, err)
	}

	if len(statements) == 0 {
		return fmt.Errorf("no executable statements found in migration %s", filename)
	}

	for _, stmt := range statements {
		log.Printf("Executing migration %s statement:\n%s", filename, stmt)
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	// Record migration in history
	// This marks the migration as applied
	_, err = tx.Exec(
		"INSERT INTO migration_history (migration_name) VALUES (?)",
		filename,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	// All changes are now permanent
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// parseSQLStatements splits a SQL file into executable statements,
// removing comments and blank lines.
func parseSQLStatements(sql string) ([]string, error) {
	scanner := bufio.NewScanner(strings.NewReader(sql))
	var statements []string
	var builder strings.Builder

	flushStatement := func() {
		stmt := strings.TrimSpace(builder.String())
		if stmt == "" {
			return
		}
		if strings.HasSuffix(stmt, ";") {
			stmt = strings.TrimSpace(strings.TrimSuffix(stmt, ";"))
		}
		if stmt != "" {
			statements = append(statements, stmt)
		}
		builder.Reset()
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		builder.WriteString(line)
		builder.WriteString("\n")

		if strings.HasSuffix(trimmed, ";") {
			flushStatement()
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	flushStatement()
	return statements, nil
}
