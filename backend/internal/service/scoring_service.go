// scoring_service.go - Content scoring business logic
// Implements the scoring algorithm for content ranking
package service

import (
	"fmt"
	"log"
	"search-engine/backend/internal/repository"
	"search-engine/backend/internal/scoring"
)

// ScoringService handles scoring operations for content
// This service orchestrates scoring calculations and database updates
type ScoringService struct {
	contentRepo *repository.ContentRepository
}

// NewScoringService creates a new ScoringService instance
func NewScoringService(contentRepo *repository.ContentRepository) *ScoringService {
	return &ScoringService{
		contentRepo: contentRepo,
	}
}

// CalculateScoreForContent calculates and updates the score for a single content item
// This is used when content is created or updated
func (s *ScoringService) CalculateScoreForContent(contentID int64) error {
	// Get content from database
	content, err := s.contentRepo.GetByID(contentID)
	if err != nil {
		return fmt.Errorf("failed to get content: %w", err)
	}

	// Calculate final score
	score := scoring.CalculateFinalScore(content)

	// Update score in database
	if err := s.contentRepo.UpdateScore(contentID, score); err != nil {
		return fmt.Errorf("failed to update score: %w", err)
	}

	log.Printf("Updated score for content %d: %.4f", contentID, score)
	return nil
}

// RecalculateAllScores recalculates scores for all content items
// This is useful when the scoring algorithm changes or for maintenance
func (s *ScoringService) RecalculateAllScores() error {
	log.Println("Starting score recalculation for all content...")

	// Get all content items in batches to avoid memory issues
	// We'll use a simple approach: get all providers and iterate through them
	// For now, we'll use a workaround by getting content with a high limit
	// In production, you might want to add a GetAllContent method to repository
	batchSize := 100
	offset := 0

	// Use a large provider ID range or implement a better method
	// For simplicity, we'll fetch from provider 1 first, then 2, etc.
	// This is a temporary solution - in production, add GetAllContent() method
	for providerID := 1; providerID <= 10; providerID++ {
		for {
			// Fetch a batch of content items
			contents, err := s.contentRepo.GetByProviderID(providerID, batchSize, offset)
			if err != nil {
				// Provider might not exist, skip to next
				break
			}

			// If no more content, we're done
			if len(contents) == 0 {
				break
			}

			// Calculate and update scores for this batch
			updated := 0
			for _, content := range contents {
				score := scoring.CalculateFinalScore(content)
				if err := s.contentRepo.UpdateScore(content.ID, score); err != nil {
					log.Printf("Failed to update score for content %d: %v", content.ID, err)
					continue
				}
				updated++
			}

			log.Printf("Updated scores for %d content items (offset: %d)", updated, offset)

			// Move to next batch
			offset += batchSize

			// If we got fewer items than batch size, move to next provider
			if len(contents) < batchSize {
				offset = 0
				break
			}
			offset += batchSize
		}
		offset = 0 // Reset for next provider
	}

	log.Println("Score recalculation completed")
	return nil
}

// RecalculateScoresForProvider recalculates scores for all content from a specific provider
// This is useful after syncing data from a provider
func (s *ScoringService) RecalculateScoresForProvider(providerID int) error {
	log.Printf("Starting score recalculation for provider %d...", providerID)

	batchSize := 100
	offset := 0

	for {
		// Fetch a batch of content items for this provider
		contents, err := s.contentRepo.GetByProviderID(providerID, batchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to get content batch: %w", err)
		}

		// If no more content, we're done
		if len(contents) == 0 {
			break
		}

		// Calculate and update scores for this batch
		updated := 0
		for _, content := range contents {
			score := scoring.CalculateFinalScore(content)
			if err := s.contentRepo.UpdateScore(content.ID, score); err != nil {
				log.Printf("Failed to update score for content %d: %v", content.ID, err)
				continue
			}
			updated++
		}

		log.Printf("Updated scores for %d content items from provider %d (offset: %d)", updated, providerID, offset)

		// Move to next batch
		offset += batchSize

		// If we got fewer items than batch size, we're done
		if len(contents) < batchSize {
			break
		}
	}

	log.Printf("Score recalculation completed for provider %d", providerID)
	return nil
}
