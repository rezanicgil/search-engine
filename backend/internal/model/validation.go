// validation.go - Model validation helpers
// Provides validation functions for models
package model

import (
	"errors"
	"fmt"
	"strings"
)

// ValidateContent validates a Content model
// Ensures all required fields are present and valid
func ValidateContent(c *Content) error {
	if c.Title == "" {
		return errors.New("title is required")
	}

	if c.Type != ContentTypeVideo && c.Type != ContentTypeArticle {
		return errors.New("type must be 'video' or 'article'")
	}

	if c.ProviderID <= 0 {
		return errors.New("provider_id must be greater than 0")
	}

	if c.ExternalID == "" {
		return errors.New("external_id is required")
	}

	if c.PublishedAt.IsZero() {
		return errors.New("published_at is required")
	}

	// Validate video-specific fields
	if c.IsVideo() {
		if c.DurationSeconds != nil && *c.DurationSeconds < 0 {
			return errors.New("duration_seconds must be non-negative")
		}
		if c.Views < 0 {
			return errors.New("views must be non-negative")
		}
		if c.Likes < 0 {
			return errors.New("likes must be non-negative")
		}
	}

	// Validate article-specific fields
	if c.IsArticle() {
		if c.ReadingTime != nil && *c.ReadingTime < 0 {
			return errors.New("reading_time must be non-negative")
		}
		if c.Reactions < 0 {
			return errors.New("reactions must be non-negative")
		}
		if c.Comments < 0 {
			return errors.New("comments must be non-negative")
		}
	}

	return nil
}

// ValidateProvider validates a Provider model
// Ensures all required fields are present and valid
func ValidateProvider(p *Provider) error {
	if p.Name == "" {
		return errors.New("name is required")
	}

	if p.URL == "" {
		return errors.New("url is required")
	}

	if !strings.HasPrefix(p.URL, "http://") && !strings.HasPrefix(p.URL, "https://") {
		return errors.New("url must start with http:// or https://")
	}

	if p.Format != ProviderFormatJSON && p.Format != ProviderFormatXML {
		return errors.New("format must be 'json' or 'xml'")
	}

	if p.RateLimitPerMinute < 1 {
		return errors.New("rate_limit_per_minute must be at least 1")
	}

	return nil
}

// NormalizeContentType normalizes a content type string
// Converts various input formats to standard ContentType
func NormalizeContentType(s string) ContentType {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "video" {
		return ContentTypeVideo
	}
	if s == "article" || s == "text" || s == "post" {
		return ContentTypeArticle
	}
	return ContentTypeVideo // Default fallback
}

// NormalizeProviderFormat normalizes a provider format string
// Converts various input formats to standard ProviderFormat
func NormalizeProviderFormat(s string) ProviderFormat {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "json" {
		return ProviderFormatJSON
	}
	if s == "xml" {
		return ProviderFormatXML
	}
	return ProviderFormatJSON // Default fallback
}

// ParseDuration parses a duration string (e.g., "15:30") to seconds
// Returns the duration in seconds and any error
func ParseDuration(durationStr string) (*int, error) {
	if durationStr == "" {
		return nil, nil
	}

	parts := strings.Split(durationStr, ":")
	if len(parts) != 2 {
		return nil, errors.New("invalid duration format, expected MM:SS or HH:MM:SS")
	}

	var totalSeconds int
	if len(parts) == 2 {
		// MM:SS format
		var minutes, seconds int
		if _, err := fmt.Sscanf(durationStr, "%d:%d", &minutes, &seconds); err != nil {
			return nil, err
		}
		totalSeconds = minutes*60 + seconds
	}

	return &totalSeconds, nil
}
