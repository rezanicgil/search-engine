// xml_provider.go - XML provider implementation (Provider 2)
// Handles fetching and parsing XML format from provider2
package provider

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"search-engine/backend/internal/model"
	"strconv"
	"strings"
	"time"
)

// XMLProviderResponse represents the XML structure from Provider 2
// XML tags match the actual API response structure
type XMLProviderResponse struct {
	XMLName xml.Name         `xml:"feed"`
	Items   []XMLContentItem `xml:"items>item"`
	Meta    XMLMeta          `xml:"meta"`
}

// XMLContentItem represents a single content item in XML format
// XML tags match the provider's XML structure
type XMLContentItem struct {
	XMLName         xml.Name      `xml:"item"`
	ID              string        `xml:"id"`
	Headline        string        `xml:"headline"`
	Type            string        `xml:"type"`
	Stats           XMLStats      `xml:"stats"`
	PublicationDate string        `xml:"publication_date"`
	Categories      XMLCategories `xml:"categories"`
}

// XMLStats represents metrics in XML format
// Can contain different metrics for videos vs articles
type XMLStats struct {
	Views    *string `xml:"views,omitempty"`    // Video metric (string in XML)
	Likes    *string `xml:"likes,omitempty"`    // Video metric (string in XML)
	Duration *string `xml:"duration,omitempty"` // Video metric (format: "MM:SS")

	ReadingTime *string `xml:"reading_time,omitempty"` // Article metric (string in XML)
	Reactions   *string `xml:"reactions,omitempty"`    // Article metric (string in XML)
	Comments    *string `xml:"comments,omitempty"`     // Article metric (string in XML)
}

// XMLCategories represents categories/tags in XML format
type XMLCategories struct {
	Category []string `xml:"category"`
}

// XMLMeta represents metadata in XML response
type XMLMeta struct {
	TotalCount   int `xml:"total_count"`
	CurrentPage  int `xml:"current_page"`
	ItemsPerPage int `xml:"items_per_page"`
}

// XMLProvider implements the Provider interface for XML format
// This handles fetching and parsing data from Provider 2
type XMLProvider struct {
	BaseProvider
	client *http.Client
}

// NewXMLProvider creates a new XML provider instance
// Sets up HTTP client with timeout for reliable requests
func NewXMLProvider(name, url string) *XMLProvider {
	return &XMLProvider{
		BaseProvider: BaseProvider{
			Name: name,
			URL:  url,
		},
		client: &http.Client{
			Timeout: 30 * time.Second, // 30 second timeout for API requests
		},
	}
}

// Fetch retrieves content from the XML provider's API
// Downloads XML data, parses it, and transforms it to standard format
func (p *XMLProvider) Fetch() ([]*model.Content, error) {
	// Make HTTP GET request to provider URL
	// This fetches the raw XML data
	resp, err := p.client.Get(p.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from XML provider: %w", err)
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

	// Parse XML response
	var xmlResponse XMLProviderResponse
	if err := xml.Unmarshal(body, &xmlResponse); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Transform XML items to standard Content models
	contents := make([]*model.Content, 0, len(xmlResponse.Items))
	for _, item := range xmlResponse.Items {
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

// transformToContent converts an XMLContentItem to a standard Content model
// This handles the transformation from provider-specific format to our standard format
func (p *XMLProvider) transformToContent(item XMLContentItem) (*model.Content, error) {
	content := &model.Content{
		ExternalID: item.ID,
		Title:      item.Headline, // XML uses "headline" instead of "title"
		Type:       model.NormalizeContentType(item.Type),
	}

	// Parse publication_date timestamp
	// XML provider uses date format: "2024-03-15"
	publishedAt, err := time.Parse("2006-01-02", item.PublicationDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse publication_date: %w", err)
	}
	content.PublishedAt = publishedAt

	// Transform metrics based on content type
	if content.IsVideo() {
		// Video metrics (XML stores numbers as strings)
		if item.Stats.Views != nil {
			views, err := parseIntString(*item.Stats.Views)
			if err != nil {
				return nil, fmt.Errorf("failed to parse views: %w", err)
			}
			content.Views = views
		}
		if item.Stats.Likes != nil {
			likes, err := parseIntString(*item.Stats.Likes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse likes: %w", err)
			}
			content.Likes = likes
		}
		if item.Stats.Duration != nil {
			// Parse duration string (e.g., "25:15") to seconds
			durationSeconds, err := parseDurationString(*item.Stats.Duration)
			if err != nil {
				return nil, fmt.Errorf("failed to parse duration: %w", err)
			}
			content.DurationSeconds = &durationSeconds
		}
	} else if content.IsArticle() {
		// Article metrics (XML stores numbers as strings)
		if item.Stats.ReadingTime != nil {
			readingTime, err := parseIntString(*item.Stats.ReadingTime)
			if err != nil {
				return nil, fmt.Errorf("failed to parse reading_time: %w", err)
			}
			content.ReadingTime = &readingTime
		}
		if item.Stats.Reactions != nil {
			reactions, err := parseIntString(*item.Stats.Reactions)
			if err != nil {
				return nil, fmt.Errorf("failed to parse reactions: %w", err)
			}
			content.Reactions = reactions
		}
		if item.Stats.Comments != nil {
			comments, err := parseIntString(*item.Stats.Comments)
			if err != nil {
				return nil, fmt.Errorf("failed to parse comments: %w", err)
			}
			content.Comments = comments
		}
	}

	// Store tags from categories (will be saved separately in content_tags table)
	content.Tags = item.Categories.Category

	return content, nil
}

// parseIntString parses a string to integer
// XML often stores numbers as strings, so we need to convert them
func parseIntString(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}
