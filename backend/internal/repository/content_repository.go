// content_repository.go - Database operations for content
// Handles all MySQL queries and data persistence
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	apperrors "search-engine/backend/internal/errors"
	"search-engine/backend/internal/model"
	"strings"
	"time"
)

// ErrContentNotFound is kept for backward compatibility
// Use apperrors.ErrContentNotFound instead
var ErrContentNotFound = apperrors.ErrContentNotFound

// ContentRepository handles all database operations for content
// This repository encapsulates content-related database queries including search
type ContentRepository struct {
	db                *sql.DB
	minFullTextLength int
}

// NewContentRepository creates a new ContentRepository instance
// minFullTextLength controls when to switch between FULLTEXT and LIKE search
func NewContentRepository(db *sql.DB, minFullTextLength int) *ContentRepository {
	if minFullTextLength <= 0 {
		minFullTextLength = 3
	}
	return &ContentRepository{
		db:                db,
		minFullTextLength: minFullTextLength,
	}
}

// Create inserts a new content item into the database
// Returns the created content with its generated ID
func (r *ContentRepository) Create(c *model.Content) error {
	// Validate content before inserting
	if err := model.ValidateContent(c); err != nil {
		return apperrors.NewValidationErrorWithDetails("Content validation failed", err.Error())
	}

	query := `
		INSERT INTO contents (
			provider_id, external_id, title, type,
			views, likes, duration_seconds,
			reading_time, reactions, comments,
			published_at, score
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(
		query,
		c.ProviderID,
		c.ExternalID,
		c.Title,
		c.Type,
		c.Views,
		c.Likes,
		c.DurationSeconds,
		c.ReadingTime,
		c.Reactions,
		c.Comments,
		c.PublishedAt,
		c.Score,
	)
	if err != nil {
		return fmt.Errorf("failed to create content: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	c.ID = id
	return nil
}

// GetByID retrieves a content item by its ID
// Returns sql.ErrNoRows if content is not found
// ctx is used for timeout and cancellation support
func (r *ContentRepository) GetByID(ctx context.Context, id int64) (*model.Content, error) {
	query := `
		SELECT id, provider_id, external_id, title, type,
		       views, likes, duration_seconds,
		       reading_time, reactions, comments,
		       published_at, score, created_at, updated_at
		FROM contents
		WHERE id = ?
	`
	c := &model.Content{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.ProviderID,
		&c.ExternalID,
		&c.Title,
		&c.Type,
		&c.Views,
		&c.Likes,
		&c.DurationSeconds,
		&c.ReadingTime,
		&c.Reactions,
		&c.Comments,
		&c.PublishedAt,
		&c.Score,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrContentNotFound
		}
		return nil, apperrors.NewDatabaseError("get content by id", err)
	}

	return c, nil
}

// GetByProviderAndExternalID retrieves content by provider ID and external ID
// This is used to check if content already exists before inserting
func (r *ContentRepository) GetByProviderAndExternalID(providerID int, externalID string) (*model.Content, error) {
	query := `
		SELECT id, provider_id, external_id, title, type,
		       views, likes, duration_seconds,
		       reading_time, reactions, comments,
		       published_at, score, created_at, updated_at
		FROM contents
		WHERE provider_id = ? AND external_id = ?
	`
	c := &model.Content{}

	err := r.db.QueryRow(query, providerID, externalID).Scan(
		&c.ID,
		&c.ProviderID,
		&c.ExternalID,
		&c.Title,
		&c.Type,
		&c.Views,
		&c.Likes,
		&c.DurationSeconds,
		&c.ReadingTime,
		&c.Reactions,
		&c.Comments,
		&c.PublishedAt,
		&c.Score,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrContentNotFound
		}
		return nil, apperrors.NewDatabaseError("get content by provider and external id", err)
	}

	return c, nil
}

// Update updates an existing content item
// Updates all fields except ID and timestamps
func (r *ContentRepository) Update(c *model.Content) error {
	// Validate content before updating
	if err := model.ValidateContent(c); err != nil {
		return apperrors.NewValidationErrorWithDetails("Content validation failed", err.Error())
	}

	query := `
		UPDATE contents
		SET title = ?, type = ?,
		    views = ?, likes = ?, duration_seconds = ?,
		    reading_time = ?, reactions = ?, comments = ?,
		    published_at = ?, score = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.Exec(
		query,
		c.Title,
		c.Type,
		c.Views,
		c.Likes,
		c.DurationSeconds,
		c.ReadingTime,
		c.Reactions,
		c.Comments,
		c.PublishedAt,
		c.Score,
		c.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update content: %w", err)
	}
	return nil
}

// UpdateScore updates only the score field for a content item
// This is used by the scoring service to update scores efficiently
func (r *ContentRepository) UpdateScore(id int64, score float64) error {
	query := `
		UPDATE contents
		SET score = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.Exec(query, score, id)
	if err != nil {
		return fmt.Errorf("failed to update score: %w", err)
	}
	return nil
}

// Delete removes a content item from the database
// This will cascade delete associated tags due to foreign key constraint
func (r *ContentRepository) Delete(id int64) error {
	query := `DELETE FROM contents WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete content: %w", err)
	}
	return nil
}

// Upsert creates or updates a content item
// If content exists (by provider_id + external_id), it updates; otherwise creates new
// This is useful when syncing data from providers
func (r *ContentRepository) Upsert(c *model.Content) error {
	existing, err := r.GetByProviderAndExternalID(c.ProviderID, c.ExternalID)
	if err != nil {
		if errors.Is(err, ErrContentNotFound) || errors.Is(err, apperrors.ErrContentNotFound) {
			return r.Create(c)
		}
		return apperrors.NewDatabaseError("check existing content", err)
	}

	c.ID = existing.ID
	return r.Update(c)
}

// Search searches for content based on the search request
// Supports keyword search, type filtering, sorting, and pagination
// ctx is used for timeout and cancellation support
func (r *ContentRepository) Search(ctx context.Context, req *model.SearchRequest) ([]*model.Content, int, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	trimmedQuery := strings.TrimSpace(req.Query)
	useFullText := len(trimmedQuery) >= r.minFullTextLength

	// Keyword search using FULLTEXT index
	if req.Query != "" {
		if useFullText {
			whereClauses = append(whereClauses, "MATCH(title) AGAINST(? IN BOOLEAN MODE)")
			args = append(args, trimmedQuery+"*")
		} else {
			whereClauses = append(whereClauses, "title LIKE ?")
			args = append(args, "%"+trimmedQuery+"%")
		}
	}

	// Type filter
	if req.Type != nil {
		whereClauses = append(whereClauses, "type = ?")
		args = append(args, *req.Type)
	}

	// Provider filter
	if req.ProviderID != nil {
		whereClauses = append(whereClauses, "provider_id = ?")
		args = append(args, *req.ProviderID)
	}

	// Date range filters
	if req.StartDate != nil {
		whereClauses = append(whereClauses, "published_at >= ?")
		args = append(args, *req.StartDate)
	}
	if req.EndDate != nil {
		whereClauses = append(whereClauses, "published_at <= ?")
		args = append(args, *req.EndDate)
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Build ORDER BY clause with whitelist validation to prevent SQL injection
	validSortFields := map[string]bool{
		"score":        true,
		"published_at": true,
		"title":        true,
		"id":           true,
	}
	sortBy := req.SortBy
	if !validSortFields[sortBy] {
		sortBy = "score" // Default to score if invalid
	}

	validSortOrders := map[string]bool{
		"ASC":  true,
		"DESC": true,
	}
	sortOrder := strings.ToUpper(req.SortOrder)
	if !validSortOrders[sortOrder] {
		sortOrder = "DESC" // Default to DESC if invalid
	}

	orderBy := fmt.Sprintf("ORDER BY %s %s, id DESC", sortBy, sortOrder)

	// Count total results (for pagination)
	// Use a separate context with timeout for COUNT query to prevent it from blocking too long
	// COUNT can be slow on large tables, so we give it a reasonable timeout
	countCtx, countCancel := context.WithTimeout(ctx, 10*time.Second)
	defer countCancel()

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM contents %s", whereClause)
	var total int
	err := r.db.QueryRowContext(countCtx, countQuery, args...).Scan(&total)
	if err != nil {
		if countCtx.Err() == context.DeadlineExceeded {
			// If COUNT times out, estimate total based on returned results
			// This allows pagination to work even if COUNT is slow
			total = -1 // Use -1 to indicate estimated/unknown total
		} else {
			return nil, 0, apperrors.NewDatabaseError("count results", err)
		}
	}

	// Build SELECT query with pagination
	query := fmt.Sprintf(`
		SELECT id, provider_id, external_id, title, type,
		       views, likes, duration_seconds,
		       reading_time, reactions, comments,
		       published_at, score, created_at, updated_at
		FROM contents
		%s
		%s
		LIMIT ? OFFSET ?
	`, whereClause, orderBy)

	args = append(args, req.PerPage, req.GetOffset())

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, apperrors.NewDatabaseError("search content", err)
	}
	defer rows.Close()

	var contents []*model.Content
	for rows.Next() {
		c := &model.Content{}
		err := rows.Scan(
			&c.ID,
			&c.ProviderID,
			&c.ExternalID,
			&c.Title,
			&c.Type,
			&c.Views,
			&c.Likes,
			&c.DurationSeconds,
			&c.ReadingTime,
			&c.Reactions,
			&c.Comments,
			&c.PublishedAt,
			&c.Score,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan content: %w", err)
		}
		contents = append(contents, c)
	}

	return contents, total, rows.Err()
}

// GetByProviderID retrieves all content items for a specific provider
// Useful for syncing or listing provider-specific content
func (r *ContentRepository) GetByProviderID(providerID int, limit, offset int) ([]*model.Content, error) {
	query := `
		SELECT id, provider_id, external_id, title, type,
		       views, likes, duration_seconds,
		       reading_time, reactions, comments,
		       published_at, score, created_at, updated_at
		FROM contents
		WHERE provider_id = ?
		ORDER BY published_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(query, providerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get content by provider id: %w", err)
	}
	defer rows.Close()

	var contents []*model.Content
	for rows.Next() {
		c := &model.Content{}
		err := rows.Scan(
			&c.ID,
			&c.ProviderID,
			&c.ExternalID,
			&c.Title,
			&c.Type,
			&c.Views,
			&c.Likes,
			&c.DurationSeconds,
			&c.ReadingTime,
			&c.Reactions,
			&c.Comments,
			&c.PublishedAt,
			&c.Score,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content: %w", err)
		}
		contents = append(contents, c)
	}

	return contents, rows.Err()
}

// LoadTags loads tags for a content item
// This is a helper method to populate the Tags field
func (r *ContentRepository) LoadTags(content *model.Content) error {
	tagRepo := NewContentTagRepository(r.db)
	tags, err := tagRepo.GetByContentID(content.ID)
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}

	content.Tags = make([]string, len(tags))
	for i, tag := range tags {
		content.Tags[i] = tag.Tag
	}

	return nil
}

// LoadTagsBatch loads tags for multiple content items efficiently
// This reduces the number of database queries when loading multiple contents
// ctx is used for timeout and cancellation support
func (r *ContentRepository) LoadTagsBatch(ctx context.Context, contents []*model.Content) error {
	if len(contents) == 0 {
		return nil
	}

	// Get all content IDs
	contentIDs := make([]int64, len(contents))
	for i, c := range contents {
		contentIDs[i] = c.ID
	}

	// Build query with IN clause
	placeholders := strings.Repeat("?,", len(contentIDs))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf(`
		SELECT content_id, tag
		FROM content_tags
		WHERE content_id IN (%s)
		ORDER BY content_id, tag
	`, placeholders)

	// Convert []int64 to []interface{}
	args := make([]interface{}, len(contentIDs))
	for i, id := range contentIDs {
		args[i] = id
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to load tags batch: %w", err)
	}
	defer rows.Close()

	// Create a map of content_id -> tags
	tagsMap := make(map[int64][]string)
	for rows.Next() {
		var contentID int64
		var tag string
		if err := rows.Scan(&contentID, &tag); err != nil {
			return fmt.Errorf("failed to scan tag: %w", err)
		}
		tagsMap[contentID] = append(tagsMap[contentID], tag)
	}

	// Populate tags in contents
	for _, c := range contents {
		if tags, ok := tagsMap[c.ID]; ok {
			c.Tags = tags
		} else {
			c.Tags = []string{} // Empty slice instead of nil
		}
	}

	return rows.Err()
}

// GetTagsByContentID retrieves all tags for a specific content item
// Returns an empty slice if no tags exist
// ctx is used for timeout and cancellation support
func (r *ContentRepository) GetTagsByContentID(ctx context.Context, contentID int64) ([]string, error) {
	query := `
		SELECT tag
		FROM content_tags
		WHERE content_id = ?
		ORDER BY tag
	`
	rows, err := r.db.QueryContext(ctx, query, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags by content id: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// GetStats retrieves statistics about the content in the database
// Returns counts by type, total count, and other useful metrics
func (r *ContentRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total content count
	var totalCount int
	err := r.db.QueryRow("SELECT COUNT(*) FROM contents").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total_content"] = totalCount

	// Count by type
	var videoCount, articleCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM contents WHERE type = 'video'").Scan(&videoCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get video count: %w", err)
	}
	stats["videos"] = videoCount

	err = r.db.QueryRow("SELECT COUNT(*) FROM contents WHERE type = 'article'").Scan(&articleCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get article count: %w", err)
	}
	stats["articles"] = articleCount

	// Count by provider
	type ProviderCount struct {
		ProviderID int    `json:"provider_id"`
		Count      int    `json:"count"`
		Name       string `json:"name,omitempty"`
	}
	var providerCounts []ProviderCount
	query := `
		SELECT provider_id, COUNT(*) as count
		FROM contents
		GROUP BY provider_id
		ORDER BY count DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pc ProviderCount
		if err := rows.Scan(&pc.ProviderID, &pc.Count); err != nil {
			return nil, fmt.Errorf("failed to scan provider count: %w", err)
		}
		providerCounts = append(providerCounts, pc)
	}
	stats["by_provider"] = providerCounts

	// Average score
	var avgScore sql.NullFloat64
	err = r.db.QueryRow("SELECT AVG(score) FROM contents").Scan(&avgScore)
	if err != nil {
		return nil, fmt.Errorf("failed to get average score: %w", err)
	}
	if avgScore.Valid {
		stats["average_score"] = avgScore.Float64
	} else {
		stats["average_score"] = 0.0
	}

	// Total tags count
	var totalTags int
	err = r.db.QueryRow("SELECT COUNT(DISTINCT tag) FROM content_tags").Scan(&totalTags)
	if err != nil {
		// Tags might not exist, so this is not critical
		totalTags = 0
	}
	stats["total_tags"] = totalTags

	return stats, nil
}
