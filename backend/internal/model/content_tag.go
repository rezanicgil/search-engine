// content_tag.go - Content tag data models
// Defines the structure for content tags (many-to-many relationship)
package model

import (
	"time"
)

// ContentTag represents a tag associated with content
// This matches the database schema in the content_tags table
// Tags are used for filtering and categorization
type ContentTag struct {
	ID        int64     `json:"id" db:"id"`
	ContentID int64     `json:"content_id" db:"content_id"`
	Tag       string    `json:"tag" db:"tag"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
