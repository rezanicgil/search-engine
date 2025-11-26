// base_score.go - Base score calculation
// Calculates base scores for video and text content types
package scoring

import (
	"search-engine/backend/internal/model"
)

// CalculateBaseScore calculates the base score for content
// Formula:
//
//	Video: views / 1000 + (likes / 100)
//	Article: reading_time + (reactions / 50)
func CalculateBaseScore(content *model.Content) float64 {
	if content.IsVideo() {
		return calculateVideoBaseScore(content)
	} else if content.IsArticle() {
		return calculateArticleBaseScore(content)
	}
	return 0.0
}

// calculateVideoBaseScore calculates base score for video content
// Formula: views / 1000 + (likes / 100)
// This normalizes views and likes to a comparable scale
func calculateVideoBaseScore(content *model.Content) float64 {
	viewsScore := float64(content.Views) / 1000.0
	likesScore := float64(content.Likes) / 100.0
	return viewsScore + likesScore
}

// calculateArticleBaseScore calculates base score for article content
// Formula: reading_time + (reactions / 50)
// Reading time is already in minutes, reactions are normalized
func calculateArticleBaseScore(content *model.Content) float64 {
	readingTimeScore := 0.0
	if content.ReadingTime != nil {
		readingTimeScore = float64(*content.ReadingTime)
	}

	reactionsScore := float64(content.Reactions) / 50.0
	return readingTimeScore + reactionsScore
}

// GetContentTypeCoefficient returns the coefficient for content type
// Video: 1.5 (videos are weighted higher)
// Article: 1.0 (articles have standard weight)
func GetContentTypeCoefficient(contentType model.ContentType) float64 {
	if contentType == model.ContentTypeVideo {
		return 1.5
	}
	return 1.0
}
