// base_provider.go - Base interface for all providers
// Defines the common contract that all providers must implement
package provider

import (
	"search-engine/backend/internal/model"
)

// Provider defines the interface that all content providers must implement
// This allows us to work with different providers (JSON, XML, etc.) uniformly
type Provider interface {
	// Fetch retrieves content from the provider's API
	// Returns a list of standardized Content models
	Fetch() ([]*model.Content, error)

	// GetName returns the provider's identifier name
	GetName() string

	// GetURL returns the provider's API endpoint URL
	GetURL() string
}

// BaseProvider contains common fields and functionality for all providers
// This reduces code duplication across different provider implementations
type BaseProvider struct {
	Name string
	URL  string
}

// GetName returns the provider name
func (p *BaseProvider) GetName() string {
	return p.Name
}

// GetURL returns the provider URL
func (p *BaseProvider) GetURL() string {
	return p.URL
}
