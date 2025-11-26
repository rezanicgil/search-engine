// provider.go - Provider data models
// Defines provider structure and metadata
package model

import (
	"time"
)

// ProviderFormat represents the data format of a provider
// This enum ensures format type safety
type ProviderFormat string

const (
	ProviderFormatJSON ProviderFormat = "json"
	ProviderFormatXML  ProviderFormat = "xml"
)

// Provider represents a content provider
// This matches the database schema in the providers table
// Providers are external sources that supply content data
type Provider struct {
	ID                 int            `json:"id" db:"id"`
	Name               string         `json:"name" db:"name"`
	URL                string         `json:"url" db:"url"`
	Format             ProviderFormat `json:"format" db:"format"`
	RateLimitPerMinute int            `json:"rate_limit_per_minute" db:"rate_limit_per_minute"`
	LastFetchedAt      *time.Time     `json:"last_fetched_at,omitempty" db:"last_fetched_at"`
	CreatedAt          time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at" db:"updated_at"`
}

// IsJSON returns true if provider format is JSON
// Helper method for format checking
func (p *Provider) IsJSON() bool {
	return p.Format == ProviderFormatJSON
}

// IsXML returns true if provider format is XML
// Helper method for format checking
func (p *Provider) IsXML() bool {
	return p.Format == ProviderFormatXML
}
