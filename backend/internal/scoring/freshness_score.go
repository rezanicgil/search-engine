// freshness_score.go - Freshness score calculation
// Calculates time-based freshness scores (1 week, 1 month, 3 months)
package scoring

import (
	"time"
)

// CalculateFreshnessScore calculates the freshness score based on publication date
// Formula:
//
//	1 week (7 days) or newer: +5
//	1 month (30 days) or newer: +3
//	3 months (90 days) or newer: +1
//	Older than 3 months: +0
func CalculateFreshnessScore(publishedAt time.Time) float64 {
	now := time.Now()
	age := now.Sub(publishedAt)

	// Calculate age in days
	days := int(age.Hours() / 24)

	// Apply freshness scoring based on age
	if days <= 7 {
		// 1 week or newer: +5 points
		return 5.0
	} else if days <= 30 {
		// 1 month or newer: +3 points
		return 3.0
	} else if days <= 90 {
		// 3 months or newer: +1 point
		return 1.0
	}

	// Older than 3 months: +0 points
	return 0.0
}

// GetAgeInDays calculates the age of content in days
// Helper function for debugging and logging
func GetAgeInDays(publishedAt time.Time) int {
	now := time.Now()
	age := now.Sub(publishedAt)
	return int(age.Hours() / 24)
}
