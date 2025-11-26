// json_provider.go - JSON provider implementation (Provider 1)
// Handles fetching and parsing JSON format from provider1
package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"search-engine/backend/internal/model"
	"time"
)

// JSONProviderResponse represents the JSON structure from Provider 1
// This matches the actual API response format
type JSONProviderResponse struct {
	Contents   []JSONContentItem `json:"contents"`
	Pagination JSONPagination    `json:"pagination"`
}

// JSONContentItem represents a single content item in JSON format
// This is the raw format from the provider before transformation
type JSONContentItem struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	Metrics     JSONMetrics `json:"metrics"`
	PublishedAt string      `json:"published_at"`
	Tags        []string    `json:"tags"`
}

// JSONMetrics represents metrics in JSON format
// Different metrics for videos vs articles
type JSONMetrics struct {
	Views    *int    `json:"views,omitempty"`    // Video metric
	Likes    *int    `json:"likes,omitempty"`    // Video metric
	Duration *string `json:"duration,omitempty"` // Video metric (format: "MM:SS")

	ReadingTime *int `json:"reading_time,omitempty"` // Article metric
	Reactions   *int `json:"reactions,omitempty"`    // Article metric
	Comments    *int `json:"comments,omitempty"`     // Article metric
}

// JSONPagination represents pagination info from JSON provider
type JSONPagination struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// JSONProvider implements the Provider interface for JSON format
// This handles fetching and parsing data from Provider 1
type JSONProvider struct {
	BaseProvider
	client *http.Client
}

// NewJSONProvider creates a new JSON provider instance
// Sets up HTTP client with timeout for reliable requests
func NewJSONProvider(name, url string) *JSONProvider {
	return &JSONProvider{
		BaseProvider: BaseProvider{
			Name: name,
			URL:  url,
		},
		client: &http.Client{
			Timeout: 30 * time.Second, // 30 second timeout for API requests
		},
	}
}

// Fetch retrieves content from the JSON provider's API
// Downloads JSON data, parses it, and transforms it to standard format
func (p *JSONProvider) Fetch() ([]*model.Content, error) {
	// Make HTTP GET request to provider URL
	// This fetches the raw JSON data
	resp, err := p.client.Get(p.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from JSON provider: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	// Non-200 status codes indicate an error
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var jsonResponse JSONProviderResponse
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Transform JSON items to standard Content models
	contents := make([]*model.Content, 0, len(jsonResponse.Contents))
	for _, item := range jsonResponse.Contents {
		content, err := p.transformToContent(item)
		if err != nil {
			// Log error but continue processing other items
			// This ensures partial failures don't stop the entire sync
			continue
		}
		contents = append(contents, content)
	}

	return contents, nil
}

// transformToContent converts a JSONContentItem to a standard Content model
// This handles the transformation from provider-specific format to our standard format
func (p *JSONProvider) transformToContent(item JSONContentItem) (*model.Content, error) {
	content := &model.Content{
		ExternalID: item.ID,
		Title:      item.Title,
		Type:       model.NormalizeContentType(item.Type),
	}

	// Parse published_at timestamp
	// JSON provider uses ISO 8601 format: "2024-03-15T10:00:00Z"
	publishedAt, err := time.Parse(time.RFC3339, item.PublishedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse published_at: %w", err)
	}
	content.PublishedAt = publishedAt

	// Transform metrics based on content type
	if content.IsVideo() {
		// Video metrics
		if item.Metrics.Views != nil {
			content.Views = *item.Metrics.Views
		}
		if item.Metrics.Likes != nil {
			content.Likes = *item.Metrics.Likes
		}
		if item.Metrics.Duration != nil {
			// Parse duration string (e.g., "15:30") to seconds
			durationSeconds, err := parseDurationString(*item.Metrics.Duration)
			if err != nil {
				return nil, fmt.Errorf("failed to parse duration: %w", err)
			}
			content.DurationSeconds = &durationSeconds
		}
	} else if content.IsArticle() {
		// Article metrics
		if item.Metrics.ReadingTime != nil {
			content.ReadingTime = item.Metrics.ReadingTime
		}
		if item.Metrics.Reactions != nil {
			content.Reactions = *item.Metrics.Reactions
		}
		if item.Metrics.Comments != nil {
			content.Comments = *item.Metrics.Comments
		}
	}

	// Store tags (will be saved separately in content_tags table)
	content.Tags = item.Tags

	return content, nil
}

// parseDurationString parses a duration string in "MM:SS" format to seconds
// Example: "15:30" -> 930 seconds
func parseDurationString(durationStr string) (int, error) {
	var minutes, seconds int
	_, err := fmt.Sscanf(durationStr, "%d:%d", &minutes, &seconds)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: %s", durationStr)
	}
	return minutes*60 + seconds, nil
}
