// search_service.go - Business logic for search operations
// Implements search, filtering, and sorting logic
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"search-engine/backend/internal/errors"
	"search-engine/backend/internal/model"
	"search-engine/backend/internal/repository"
	"search-engine/backend/pkg/cache"
	"time"
)

// SearchService handles search operations
// This service orchestrates search queries, result processing, and response formatting
type SearchService struct {
	contentRepo        *repository.ContentRepository
	cache              cache.Cache
	cacheTTL           time.Duration
	queryTimeout       time.Duration
	simpleQueryTimeout time.Duration
}

// NewSearchService creates a new SearchService instance
// cache can be nil to disable caching.
// queryTimeout is the timeout for search queries (default: 15s)
// simpleQueryTimeout is the timeout for simple queries like GetByID (default: 5s)
func NewSearchService(contentRepo *repository.ContentRepository, cache cache.Cache, cacheTTL, queryTimeout, simpleQueryTimeout time.Duration) *SearchService {
	if cacheTTL <= 0 {
		cacheTTL = time.Minute
	}
	if queryTimeout <= 0 {
		queryTimeout = 15 * time.Second
	}
	if simpleQueryTimeout <= 0 {
		simpleQueryTimeout = 5 * time.Second
	}
	return &SearchService{
		contentRepo:        contentRepo,
		cache:              cache,
		cacheTTL:           cacheTTL,
		queryTimeout:       queryTimeout,
		simpleQueryTimeout: simpleQueryTimeout,
	}
}

// Search performs a search query and returns formatted results
// This is the main entry point for search operations
// It handles validation, searching, tag loading, and response formatting
// ctx is used for timeout and cancellation support
func (s *SearchService) Search(ctx context.Context, req *model.SearchRequest) (*model.SearchResponse, error) {
	// Validate and set default values for the request
	// This ensures we have valid parameters even if client doesn't provide them
	req.Validate()

	cacheKey := ""
	if s.cache != nil {
		cacheKey = buildSearchCacheKey(req)
		if cached, ok := s.cache.Get(cacheKey); ok {
			switch v := cached.(type) {
			case *model.SearchResponse:
				return v, nil
			case []byte:
				var resp model.SearchResponse
				if err := json.Unmarshal(v, &resp); err == nil {
					return &resp, nil
				}
			}
		}
	}

	// Apply timeout for search query (longer timeout for complex searches)
	searchCtx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	// Perform the search using the repository
	// The repository handles the actual database query with filtering and sorting
	contents, total, err := s.contentRepo.Search(searchCtx, req)
	if err != nil {
		// Check if it's already an AppError
		if appErr := errors.AsAppError(err); appErr != nil {
			return nil, appErr
		}

		if searchCtx.Err() == context.DeadlineExceeded {
			return nil, errors.NewQueryTimeoutError("search")
		}
		return nil, errors.NewServiceError("search content", err)
	}

	// Load tags for all content items in batch
	// This is more efficient than loading tags one by one
	// Use shorter timeout for tag loading (simpler query)
	if len(contents) > 0 {
		tagCtx, tagCancel := context.WithTimeout(ctx, s.simpleQueryTimeout)
		if err := s.contentRepo.LoadTagsBatch(tagCtx, contents); err != nil {
			// Log error but don't fail the entire search
			// Tags are optional metadata
			if tagCtx.Err() == context.DeadlineExceeded {
				fmt.Printf("Warning: tag loading timeout after %v\n", s.simpleQueryTimeout)
			} else {
				fmt.Printf("Warning: failed to load tags: %v\n", err)
			}
		}
		tagCancel()
	}

	// Convert repository results to response format
	// We need to convert []*model.Content to []model.Content for JSON serialization
	results := make([]model.Content, len(contents))
	for i, content := range contents {
		results[i] = *content
	}

	// Build the search response
	response := &model.SearchResponse{
		Results: results,
		Total:   total,
		Page:    req.Page,
		PerPage: req.PerPage,
	}

	// Calculate total pages for pagination metadata
	// This helps clients build pagination UI
	response.CalculateTotalPages()

	// Store in cache for subsequent requests
	if s.cache != nil && cacheKey != "" {
		// For RedisCache we pass JSON bytes; InMemoryCache will also accept []byte.
		if b, err := json.Marshal(response); err == nil {
			s.cache.Set(cacheKey, b, s.cacheTTL)
		} else {
			// Fallback: store as pointer for in-memory cache if JSON fails.
			s.cache.Set(cacheKey, response, s.cacheTTL)
		}
	}

	return response, nil
}

// buildSearchCacheKey builds a cache key that uniquely identifies a search request.
func buildSearchCacheKey(r *model.SearchRequest) string {
	// We keep it simple and explicit instead of generic JSON serialization.
	key := fmt.Sprintf("q=%s|t=%s|p=%d|prov=%v|sd=%v|ed=%v|sort=%s|ord=%s|pp=%d",
		r.Query,
		func() string {
			if r.Type == nil {
				return ""
			}
			return string(*r.Type)
		}(),
		r.Page,
		func() int {
			if r.ProviderID == nil {
				return 0
			}
			return *r.ProviderID
		}(),
		r.StartDate,
		r.EndDate,
		r.SortBy,
		r.SortOrder,
		r.PerPage,
	)
	return key
}
