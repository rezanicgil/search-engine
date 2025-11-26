// content_tag_repository.go - Database operations for content tags
// Handles tag-related database queries
package repository

import (
	"database/sql"
	"fmt"
	"search-engine/backend/internal/model"
)

// ContentTagRepository handles all database operations for content tags
// This repository encapsulates tag-related database queries
type ContentTagRepository struct {
	db *sql.DB
}

// NewContentTagRepository creates a new ContentTagRepository instance
// This allows dependency injection of the database connection
func NewContentTagRepository(db *sql.DB) *ContentTagRepository {
	return &ContentTagRepository{db: db}
}

// Create inserts a new tag for a content item
// Returns the created tag with its generated ID
func (r *ContentTagRepository) Create(tag *model.ContentTag) error {
	query := `
		INSERT INTO content_tags (content_id, tag)
		VALUES (?, ?)
	`
	result, err := r.db.Exec(query, tag.ContentID, tag.Tag)
	if err != nil {
		return fmt.Errorf("failed to create content tag: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	tag.ID = id
	return nil
}

// CreateBatch inserts multiple tags for a content item efficiently
// This reduces database round trips when adding multiple tags
func (r *ContentTagRepository) CreateBatch(contentID int64, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	// Build query with multiple values
	query := "INSERT INTO content_tags (content_id, tag) VALUES "
	args := make([]interface{}, 0, len(tags)*2)

	for i, tag := range tags {
		if i > 0 {
			query += ", "
		}
		query += "(?, ?)"
		args = append(args, contentID, tag)
	}

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to create content tags batch: %w", err)
	}

	return nil
}

// GetByContentID retrieves all tags for a specific content item
// Returns an empty slice if no tags exist
func (r *ContentTagRepository) GetByContentID(contentID int64) ([]*model.ContentTag, error) {
	query := `
		SELECT id, content_id, tag, created_at
		FROM content_tags
		WHERE content_id = ?
		ORDER BY tag
	`
	rows, err := r.db.Query(query, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags by content id: %w", err)
	}
	defer rows.Close()

	var tags []*model.ContentTag
	for rows.Next() {
		tag := &model.ContentTag{}
		err := rows.Scan(&tag.ID, &tag.ContentID, &tag.Tag, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// DeleteByContentID removes all tags for a specific content item
// This is useful when updating content tags
func (r *ContentTagRepository) DeleteByContentID(contentID int64) error {
	query := `DELETE FROM content_tags WHERE content_id = ?`
	_, err := r.db.Exec(query, contentID)
	if err != nil {
		return fmt.Errorf("failed to delete tags by content id: %w", err)
	}
	return nil
}

// Delete removes a specific tag from a content item
func (r *ContentTagRepository) Delete(contentID int64, tag string) error {
	query := `DELETE FROM content_tags WHERE content_id = ? AND tag = ?`
	_, err := r.db.Exec(query, contentID, tag)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	return nil
}

// ReplaceTags replaces all tags for a content item
// This is a convenience method that deletes old tags and creates new ones
func (r *ContentTagRepository) ReplaceTags(contentID int64, tags []string) error {
	// Start transaction for atomicity
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing tags
	deleteQuery := `DELETE FROM content_tags WHERE content_id = ?`
	_, err = tx.Exec(deleteQuery, contentID)
	if err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// Insert new tags if any
	if len(tags) > 0 {
		insertQuery := "INSERT INTO content_tags (content_id, tag) VALUES "
		args := make([]interface{}, 0, len(tags)*2)

		for i, tag := range tags {
			if i > 0 {
				insertQuery += ", "
			}
			insertQuery += "(?, ?)"
			args = append(args, contentID, tag)
		}

		_, err = tx.Exec(insertQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to insert new tags: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
