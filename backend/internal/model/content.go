// content.go - Content data models
// Defines the standard content structure used throughout the application
package model

import (
	"fmt"
	"time"
)

// ContentType represents the type of content (video or article)
// This enum ensures type safety throughout the application
type ContentType string

const (
	ContentTypeVideo   ContentType = "video"
	ContentTypeArticle ContentType = "article"
)

// Content represents a content item from any provider
// This is the standardized format that unifies data from different providers
// It matches the database schema in the contents table
type Content struct {
	ID         int64       `json:"id" db:"id"`
	ProviderID int         `json:"provider_id" db:"provider_id"`
	ExternalID string      `json:"external_id" db:"external_id"`
	Title      string      `json:"title" db:"title"`
	Type       ContentType `json:"type" db:"type"`

	// Video-specific metrics (NULL for articles)
	// These fields are populated when Type is "video"
	Views           int  `json:"views,omitempty" db:"views"`
	Likes           int  `json:"likes,omitempty" db:"likes"`
	DurationSeconds *int `json:"duration_seconds,omitempty" db:"duration_seconds"`

	// Article-specific metrics (NULL for videos)
	// These fields are populated when Type is "article"
	ReadingTime *int `json:"reading_time,omitempty" db:"reading_time"`
	Reactions   int  `json:"reactions,omitempty" db:"reactions"`
	Comments    int  `json:"comments,omitempty" db:"comments"`

	// Common fields
	PublishedAt time.Time `json:"published_at" db:"published_at"`
	Score       float64   `json:"score" db:"score"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Related data (loaded separately)
	Tags     []string  `json:"tags,omitempty"`     // Tags associated with this content
	Provider *Provider `json:"provider,omitempty"` // Provider information (optional)
}

// IsVideo returns true if content type is video
// Helper method for type checking
func (c *Content) IsVideo() bool {
	return c.Type == ContentTypeVideo
}

// IsArticle returns true if content type is article
// Helper method for type checking
func (c *Content) IsArticle() bool {
	return c.Type == ContentTypeArticle
}

// GetDurationString returns duration in "MM:SS" format for videos
// Returns empty string for articles or if duration is not set
func (c *Content) GetDurationString() string {
	if !c.IsVideo() || c.DurationSeconds == nil {
		return ""
	}
	minutes := *c.DurationSeconds / 60
	seconds := *c.DurationSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
