// main.go - Seed script to add test data for pagination testing
// Adds multiple content items to test pagination functionality

package main

import (
	"fmt"
	"log"
	"math/rand"
	"search-engine/backend/internal/config"
	"search-engine/backend/internal/model"
	"search-engine/backend/internal/repository"
	"search-engine/backend/internal/service"
	"time"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}

	if err := repository.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repository.Close()

	providerRepo := repository.NewProviderRepository(repository.GetDB())
	contentRepo := repository.NewContentRepository(repository.GetDB(), cfg.Search.MinFullTextLength)
	tagRepo := repository.NewContentTagRepository(repository.GetDB())

	// Get providers
	providers, err := providerRepo.GetAll()
	if err != nil || len(providers) == 0 {
		log.Fatalf("No providers found. Please run sync first.")
	}

	log.Printf("Found %d providers. Adding test data...", len(providers))

	// Generate test content
	rand.Seed(time.Now().UnixNano())
	titles := []string{
		"Introduction to Go Programming",
		"Advanced React Patterns",
		"Docker Containerization Guide",
		"Microservices Architecture",
		"Database Design Best Practices",
		"RESTful API Development",
		"GraphQL Fundamentals",
		"Kubernetes Orchestration",
		"CI/CD Pipeline Setup",
		"Cloud Computing Basics",
		"Machine Learning Basics",
		"Web Security Essentials",
		"Performance Optimization",
		"System Design Principles",
		"Agile Development Methodologies",
		"DevOps Best Practices",
		"Frontend Frameworks Comparison",
		"Backend Architecture Patterns",
		"API Gateway Design",
		"Message Queue Systems",
		"Distributed Systems Concepts",
		"Container Orchestration",
		"Serverless Architecture",
		"Data Structures and Algorithms",
		"Software Testing Strategies",
		"Code Review Best Practices",
		"Version Control with Git",
		"Database Migration Strategies",
		"API Documentation Tools",
		"Monitoring and Logging",
	}

	contentTypes := []model.ContentType{
		model.ContentTypeVideo,
		model.ContentTypeArticle,
	}

	// Add 50 test items
	totalItems := 50
	addedCount := 0

	for i := 0; i < totalItems; i++ {
		provider := providers[i%len(providers)]
		contentType := contentTypes[i%len(contentTypes)]
		titleIndex := i % len(titles)

		content := &model.Content{
			ProviderID:  provider.ID,
			ExternalID:  fmt.Sprintf("test-%d-%d", provider.ID, i+1),
			Title:       titles[titleIndex],
			Type:        contentType,
			PublishedAt: time.Now().AddDate(0, 0, -rand.Intn(365)),
		}

		if contentType == model.ContentTypeVideo {
			content.Views = rand.Intn(100000) + 1000
			content.Likes = rand.Intn(10000) + 100
			duration := rand.Intn(3600) + 60
			content.DurationSeconds = &duration
		} else {
			readingTime := rand.Intn(30) + 5
			content.ReadingTime = &readingTime
			content.Reactions = rand.Intn(5000) + 50
			content.Comments = rand.Intn(1000) + 10
		}

		// Check if content already exists
		existing, err := contentRepo.GetByProviderAndExternalID(provider.ID, content.ExternalID)
		if err == nil && existing != nil {
			log.Printf("Content %s already exists, skipping...", content.ExternalID)
			continue
		}

		// Create content
		if err := contentRepo.Create(content); err != nil {
			log.Printf("Failed to create content %s: %v", content.ExternalID, err)
			continue
		}

		// Add some tags
		tags := []string{
			"tutorial",
			"programming",
			"development",
			"technology",
		}
		selectedTags := tags[:rand.Intn(len(tags))+1]
		if err := tagRepo.CreateBatch(content.ID, selectedTags); err != nil {
			log.Printf("Failed to add tags for content %d: %v", content.ID, err)
		}

		addedCount++
		if addedCount%10 == 0 {
			log.Printf("Added %d/%d items...", addedCount, totalItems)
		}
	}

	log.Printf("Successfully added %d test content items", addedCount)

	// Recalculate scores
	log.Println("Recalculating scores...")
	scoringService := service.NewScoringService(contentRepo)
	for _, p := range providers {
		if err := scoringService.RecalculateScoresForProvider(p.ID); err != nil {
			log.Printf("Failed to recalculate scores for provider %d: %v", p.ID, err)
			continue
		}
	}
	log.Println("Score recalculation completed")
}
