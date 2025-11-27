// search.go - Search request and response models
// Defines the API models for search operations
package model

import "time"

// SearchRequest represents the search query parameters
// This is what the API receives from clients
type SearchRequest struct {
	Query      string       `json:"query,omitempty" form:"query"`                                   // Search keyword (optional - if empty, returns all content)
	Type       *ContentType `json:"type,omitempty" form:"type"`                                      // Filter by content type (optional)
	ProviderID *int         `json:"provider_id,omitempty" form:"provider_id"`                        // Filter by provider (optional)
	StartDate  *time.Time   `json:"start_date,omitempty" form:"start_date" time_format:"2006-01-02"` // Filter by published_at >= start_date
	EndDate    *time.Time   `json:"end_date,omitempty" form:"end_date" time_format:"2006-01-02"`     // Filter by published_at <= end_date
	Page       int          `json:"page,omitempty" form:"page"`                                      // Page number (default: 1)
	PerPage    int          `json:"per_page,omitempty" form:"per_page"`                              // Items per page (default: 10)
	SortBy     string       `json:"sort_by,omitempty" form:"sort_by"`                                // Sort field: "score", "published_at" (default: "score")
	SortOrder  string       `json:"sort_order,omitempty" form:"sort_order"`                          // Sort order: "asc", "desc" (default: "desc")
}

// Validate validates and sets default values for SearchRequest
// This ensures the request has valid parameters before processing
func (r *SearchRequest) Validate() {
	// Set default page
	if r.Page < 1 {
		r.Page = 1
	}

	// Set default per_page (limit to prevent excessive results)
	if r.PerPage < 1 {
		r.PerPage = 10
	}
	if r.PerPage > 100 {
		r.PerPage = 100 // Maximum limit
	}

	// Set default sort_by
	if r.SortBy == "" {
		r.SortBy = "score"
	}

	// Validate sort_by values
	validSortFields := map[string]bool{
		"score":        true,
		"published_at": true,
		"title":        true,
	}
	if !validSortFields[r.SortBy] {
		r.SortBy = "score" // Default to score if invalid
	}

	// Set default sort_order
	if r.SortOrder == "" {
		r.SortOrder = "desc"
	}

	// Validate sort_order
	if r.SortOrder != "asc" && r.SortOrder != "desc" {
		r.SortOrder = "desc" // Default to desc if invalid
	}

	// Normalize date range
	if r.StartDate != nil && r.EndDate != nil {
		if r.EndDate.Before(*r.StartDate) {
			// Swap to maintain chronological order
			start := *r.StartDate
			r.StartDate = r.EndDate
			r.EndDate = &start
		}
	}
}

// GetOffset calculates the database offset for pagination
// Used in SQL LIMIT/OFFSET queries
func (r *SearchRequest) GetOffset() int {
	return (r.Page - 1) * r.PerPage
}

// SearchResponse represents the search results
// This is what the API returns to clients
type SearchResponse struct {
	Results    []Content `json:"results"`     // Search results
	Total      int       `json:"total"`       // Total number of results
	Page       int       `json:"page"`        // Current page number
	PerPage    int       `json:"per_page"`    // Items per page
	TotalPages int       `json:"total_pages"` // Total number of pages
}

// CalculateTotalPages computes the total number of pages based on total results
// Helper method for pagination metadata
// If total is -1 (unknown/estimated), total_pages will be 0
func (r *SearchResponse) CalculateTotalPages() {
	if r.Total < 0 {
		// Total is unknown (e.g., COUNT query timed out)
		r.TotalPages = 0
		return
	}
	if r.PerPage > 0 {
		r.TotalPages = (r.Total + r.PerPage - 1) / r.PerPage // Ceiling division
	} else {
		r.TotalPages = 0
	}
}
