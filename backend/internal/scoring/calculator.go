// calculator.go - Scoring algorithm implementation
// Implements the final score calculation formula
package scoring

import (
	"search-engine/backend/internal/model"
)

// CalculateFinalScore calculates the final score for content
// Formula: (Base Score * Content Type Coefficient) + Freshness Score + Engagement Score
//
// Base Score:
//
//	Video: views / 1000 + (likes / 100)
//	Article: reading_time + (reactions / 50)
//
// Content Type Coefficient:
//
//	Video: 1.5
//	Article: 1.0
//
// Freshness Score:
//
//	1 week or newer: +5
//	1 month or newer: +3
//	3 months or newer: +1
//	Older: +0
//
// Engagement Score:
//
//	Video: (likes / views) * 10
//	Article: (reactions / reading_time) * 5
func CalculateFinalScore(content *model.Content) float64 {
	// Step 1: Calculate base score
	baseScore := CalculateBaseScore(content)

	// Step 2: Apply content type coefficient
	coefficient := GetContentTypeCoefficient(content.Type)
	weightedBaseScore := baseScore * coefficient

	// Step 3: Calculate freshness score
	freshnessScore := CalculateFreshnessScore(content.PublishedAt)

	// Step 4: Calculate engagement score
	engagementScore := CalculateEngagementScore(content)

	// Step 5: Combine all scores
	finalScore := weightedBaseScore + freshnessScore + engagementScore

	return finalScore
}

// CalculateAndUpdateScore calculates the final score and updates the content
// This is a convenience method that both calculates and sets the score
func CalculateAndUpdateScore(content *model.Content) {
	content.Score = CalculateFinalScore(content)
}
