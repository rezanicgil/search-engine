package main

import (
	"errors"
	"log"

	"search-engine/backend/internal/config"
	"search-engine/backend/internal/model"
	"search-engine/backend/internal/provider"
	"search-engine/backend/internal/repository"
	"search-engine/backend/internal/service"
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

	manager := provider.NewManager(providerRepo, contentRepo, tagRepo)

	ensureProvider(providerRepo, &model.Provider{
		Name:               "provider1",
		URL:                cfg.Provider.Provider1URL,
		Format:             model.ProviderFormatJSON,
		RateLimitPerMinute: 60,
	})

	ensureProvider(providerRepo, &model.Provider{
		Name:               "provider2",
		URL:                cfg.Provider.Provider2URL,
		Format:             model.ProviderFormatXML,
		RateLimitPerMinute: 60,
	})

	manager.RegisterProvider(provider.NewJSONProvider("provider1", cfg.Provider.Provider1URL))
	manager.RegisterProvider(provider.NewXMLProvider("provider2", cfg.Provider.Provider2URL))

	log.Println("Fetching data from providers...")
	if err := manager.FetchAll(); err != nil {
		log.Fatalf("Failed to fetch providers: %v", err)
	}

	log.Println("Provider sync completed successfully")

	// After syncing content, recalculate scores so that search ordering by score is meaningful.
	scoringService := service.NewScoringService(contentRepo)

	providers, err := providerRepo.GetAll()
	if err != nil {
		log.Fatalf("Failed to fetch providers for scoring: %v", err)
	}

	for _, p := range providers {
		if err := scoringService.RecalculateScoresForProvider(p.ID); err != nil {
			log.Printf("Failed to recalculate scores for provider %d: %v", p.ID, err)
			continue
		}
	}

	log.Println("Score recalculation for all providers completed successfully")
}

func ensureProvider(repo *repository.ProviderRepository, p *model.Provider) {
	existing, err := repo.GetByName(p.Name)
	if err != nil {
		if errors.Is(err, repository.ErrProviderNotFound) {
			if err := repo.Create(p); err != nil {
				log.Fatalf("Failed to create provider %s: %v", p.Name, err)
			}
			log.Printf("Created provider %s", p.Name)
			return
		}
		log.Fatalf("Failed to get provider %s: %v", p.Name, err)
		return
	}

	existing.URL = p.URL
	existing.Format = p.Format
	existing.RateLimitPerMinute = p.RateLimitPerMinute
	if err := repo.Update(existing); err != nil {
		log.Fatalf("Failed to update provider %s: %v", existing.Name, err)
	}
	log.Printf("Updated provider %s", existing.Name)
}
