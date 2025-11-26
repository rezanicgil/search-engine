// manager.go - Provider manager
// Orchestrates multiple providers, handles rate limiting, and data aggregation
package provider

import (
	"fmt"
	"log"
	"search-engine/backend/internal/repository"
	"sync"
	"time"
)

// Manager orchestrates multiple content providers
// Handles fetching from all providers, rate limiting, and data persistence
type Manager struct {
	providers    map[string]Provider
	providerRepo *repository.ProviderRepository
	contentRepo  *repository.ContentRepository
	tagRepo      *repository.ContentTagRepository
	rateLimiters map[string]*RateLimiter
	mu           sync.RWMutex // Protects rateLimiters map
}

// NewManager creates a new ProviderManager instance
// Initializes rate limiters for each provider
func NewManager(
	providerRepo *repository.ProviderRepository,
	contentRepo *repository.ContentRepository,
	tagRepo *repository.ContentTagRepository,
) *Manager {
	return &Manager{
		providers:    make(map[string]Provider),
		providerRepo: providerRepo,
		contentRepo:  contentRepo,
		tagRepo:      tagRepo,
		rateLimiters: make(map[string]*RateLimiter),
	}
}

// RegisterProvider adds a provider to the manager
// This allows the manager to fetch from multiple providers
func (m *Manager) RegisterProvider(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers[provider.GetName()] = provider

	// Initialize rate limiter for this provider
	// Get rate limit from database if provider exists
	providerModel, err := m.providerRepo.GetByName(provider.GetName())
	if err == nil {
		m.rateLimiters[provider.GetName()] = NewRateLimiter(providerModel.RateLimitPerMinute)
	} else {
		// Default rate limit if provider not in database
		m.rateLimiters[provider.GetName()] = NewRateLimiter(60)
	}
}

// FetchAll fetches content from all registered providers
// Handles rate limiting and error recovery per provider
func (m *Manager) FetchAll() error {
	var wg sync.WaitGroup
	errors := make(chan error, len(m.providers))

	m.mu.RLock()
	providers := make([]Provider, 0, len(m.providers))
	for _, p := range m.providers {
		providers = append(providers, p)
	}
	m.mu.RUnlock()

	// Fetch from all providers concurrently
	// Each provider runs in its own goroutine for parallel processing
	for _, provider := range providers {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			if err := m.fetchFromProvider(p); err != nil {
				log.Printf("Error fetching from provider %s: %v", p.GetName(), err)
				errors <- err
			}
		}(provider)
	}

	wg.Wait()
	close(errors)

	// Check if any errors occurred
	hasErrors := false
	for err := range errors {
		if err != nil {
			hasErrors = true
			log.Printf("Provider fetch error: %v", err)
		}
	}

	if hasErrors {
		return fmt.Errorf("some providers failed to fetch")
	}

	return nil
}

// fetchFromProvider fetches content from a single provider
// Handles rate limiting, data transformation, and database persistence
func (m *Manager) fetchFromProvider(provider Provider) error {
	providerName := provider.GetName()

	// Get rate limiter for this provider
	m.mu.RLock()
	limiter, exists := m.rateLimiters[providerName]
	m.mu.RUnlock()

	if !exists {
		limiter = NewRateLimiter(60) // Default rate limit
	}

	// Wait for rate limit before making request
	// This prevents exceeding the provider's rate limit
	limiter.Wait()

	log.Printf("Fetching from provider: %s", providerName)

	// Fetch content from provider
	contents, err := provider.Fetch()
	if err != nil {
		return fmt.Errorf("failed to fetch from provider %s: %w", providerName, err)
	}

	log.Printf("Fetched %d items from provider: %s", len(contents), providerName)

	// Get provider model from database
	providerModel, err := m.providerRepo.GetByName(providerName)
	if err != nil {
		return fmt.Errorf("provider not found in database: %s", providerName)
	}

	// Save each content item to database
	// Use Upsert to handle duplicates (same external_id from same provider)
	for _, content := range contents {
		content.ProviderID = providerModel.ID

		// Upsert content (create or update)
		if err := m.contentRepo.Upsert(content); err != nil {
			log.Printf("Failed to upsert content %s: %v", content.ExternalID, err)
			continue
		}

		// Get the content ID (needed for tags)
		existingContent, err := m.contentRepo.GetByProviderAndExternalID(
			content.ProviderID,
			content.ExternalID,
		)
		if err != nil {
			log.Printf("Failed to get content after upsert: %v", err)
			continue
		}

		// Save tags
		if len(content.Tags) > 0 {
			if err := m.tagRepo.ReplaceTags(existingContent.ID, content.Tags); err != nil {
				log.Printf("Failed to save tags for content %d: %v", existingContent.ID, err)
			}
		}
	}

	// Update last_fetched_at timestamp
	if err := m.providerRepo.UpdateLastFetched(providerModel.ID, time.Now()); err != nil {
		log.Printf("Failed to update last_fetched_at for provider %s: %v", providerName, err)
	}

	log.Printf("Successfully synced %d items from provider: %s", len(contents), providerName)
	return nil
}

// FetchFromProvider fetches content from a specific provider by name
// Useful for manual sync or testing individual providers
func (m *Manager) FetchFromProvider(providerName string) error {
	m.mu.RLock()
	provider, exists := m.providers[providerName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("provider not found: %s", providerName)
	}

	return m.fetchFromProvider(provider)
}

// GetProviders returns a list of all registered provider names
func (m *Manager) GetProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}
