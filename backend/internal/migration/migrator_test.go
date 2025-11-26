package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSQLStatements(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "001_initial_schema.sql")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}

	statements, err := parseSQLStatements(string(data))
	if err != nil {
		t.Fatalf("parseSQLStatements returned error: %v", err)
	}

	if len(statements) != 4 {
		t.Fatalf("expected 4 statements, got %d", len(statements))
	}
}
