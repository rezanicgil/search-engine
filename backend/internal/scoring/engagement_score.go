// engagement_score.go - Engagement score calculation
// Calculates user interaction scores (likes/views, reactions/reading_time)
package scoring

import (
	"search-engine/backend/internal/model"
)

// CalculateEngagementScore calculates the engagement score based on user interactions
// Formula:
//
//	Video: (likes / views) * 10
//	Article: (reactions / reading_time) * 5
func CalculateEngagementScore(content *model.Content) float64 {
	if content.IsVideo() {
		return calculateVideoEngagementScore(content)
	} else if content.IsArticle() {
		return calculateArticleEngagementScore(content)
	}
	return 0.0
}

// calculateVideoEngagementScore calculates engagement score for video content
// Formula: (likes / views) * 10
// This measures the like-to-view ratio, indicating content quality
// Returns 0 if views is 0 to avoid division by zero
func calculateVideoEngagementScore(content *model.Content) float64 {
	if content.Views == 0 {
		return 0.0
	}

	// Calculate like-to-view ratio
	ratio := float64(content.Likes) / float64(content.Views)

	// Multiply by 10 to scale the score
	return ratio * 10.0
}

// calculateArticleEngagementScore calculates engagement score for article content
// Formula: (reactions / reading_time) * 5
// This measures reactions per minute of reading time
// Returns 0 if reading_time is 0 or nil to avoid division by zero
func calculateArticleEngagementScore(content *model.Content) float64 {
	if content.ReadingTime == nil || *content.ReadingTime == 0 {
		return 0.0
	}

	// Calculate reactions per minute of reading time
	ratio := float64(content.Reactions) / float64(*content.ReadingTime)

	// Multiply by 5 to scale the score
	return ratio * 5.0
}
