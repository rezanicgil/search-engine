// provider_repository.go - Database operations for providers
// Manages provider metadata and request limits
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"search-engine/backend/internal/model"
	"time"
)

var ErrProviderNotFound = errors.New("provider not found")

// ProviderRepository handles all database operations for providers
// This repository encapsulates provider-related database queries
type ProviderRepository struct {
	db *sql.DB
}

// NewProviderRepository creates a new ProviderRepository instance
// This allows dependency injection of the database connection
func NewProviderRepository(db *sql.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

// Create inserts a new provider into the database
// Returns the created provider with its generated ID
func (r *ProviderRepository) Create(p *model.Provider) error {
	query := `
		INSERT INTO providers (name, url, format, rate_limit_per_minute)
		VALUES (?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, p.Name, p.URL, p.Format, p.RateLimitPerMinute)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	p.ID = int(id)
	return nil
}

// GetByID retrieves a provider by its ID
// Returns sql.ErrNoRows if provider is not found
func (r *ProviderRepository) GetByID(id int) (*model.Provider, error) {
	query := `
		SELECT id, name, url, format, rate_limit_per_minute, 
		       last_fetched_at, created_at, updated_at
		FROM providers
		WHERE id = ?
	`
	p := &model.Provider{}
	var lastFetchedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&p.ID,
		&p.Name,
		&p.URL,
		&p.Format,
		&p.RateLimitPerMinute,
		&lastFetchedAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProviderNotFound
		}
		return nil, fmt.Errorf("failed to get provider by id: %w", err)
	}

	if lastFetchedAt.Valid {
		p.LastFetchedAt = &lastFetchedAt.Time
	}

	return p, nil
}

// GetByName retrieves a provider by its name
// Returns sql.ErrNoRows if provider is not found
func (r *ProviderRepository) GetByName(name string) (*model.Provider, error) {
	query := `
		SELECT id, name, url, format, rate_limit_per_minute, 
		       last_fetched_at, created_at, updated_at
		FROM providers
		WHERE name = ?
	`
	p := &model.Provider{}
	var lastFetchedAt sql.NullTime

	err := r.db.QueryRow(query, name).Scan(
		&p.ID,
		&p.Name,
		&p.URL,
		&p.Format,
		&p.RateLimitPerMinute,
		&lastFetchedAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProviderNotFound
		}
		return nil, fmt.Errorf("failed to get provider by name: %w", err)
	}

	if lastFetchedAt.Valid {
		p.LastFetchedAt = &lastFetchedAt.Time
	}

	return p, nil
}

// GetAll retrieves all providers from the database
// Returns an empty slice if no providers exist
func (r *ProviderRepository) GetAll() ([]*model.Provider, error) {
	query := `
		SELECT id, name, url, format, rate_limit_per_minute, 
		       last_fetched_at, created_at, updated_at
		FROM providers
		ORDER BY name
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all providers: %w", err)
	}
	defer rows.Close()

	var providers []*model.Provider
	for rows.Next() {
		p := &model.Provider{}
		var lastFetchedAt sql.NullTime

		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.URL,
			&p.Format,
			&p.RateLimitPerMinute,
			&lastFetchedAt,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}

		if lastFetchedAt.Valid {
			p.LastFetchedAt = &lastFetchedAt.Time
		}

		providers = append(providers, p)
	}

	return providers, rows.Err()
}

// UpdateLastFetched updates the last_fetched_at timestamp for a provider
// This is used to track when data was last successfully fetched from the provider
func (r *ProviderRepository) UpdateLastFetched(id int, fetchedAt time.Time) error {
	query := `
		UPDATE providers
		SET last_fetched_at = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.Exec(query, fetchedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update last fetched: %w", err)
	}
	return nil
}

// Update updates provider information
// Only updates non-zero fields
func (r *ProviderRepository) Update(p *model.Provider) error {
	query := `
		UPDATE providers
		SET name = ?, url = ?, format = ?, rate_limit_per_minute = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.Exec(query, p.Name, p.URL, p.Format, p.RateLimitPerMinute, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update provider: %w", err)
	}
	return nil
}

// Delete removes a provider from the database
// Note: This will cascade delete all associated contents due to foreign key constraint
func (r *ProviderRepository) Delete(id int) error {
	query := `DELETE FROM providers WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}
	return nil
}
