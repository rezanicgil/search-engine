// main.go - Main entry point for the Search Engine API server
// Initializes database, services, handlers, and starts the HTTP server
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"search-engine/backend/internal/config"
	"search-engine/backend/internal/handler"
	"search-engine/backend/internal/middleware"
	"search-engine/backend/internal/migration"
	"search-engine/backend/internal/model"
	"search-engine/backend/internal/provider"
	"search-engine/backend/internal/repository"
	"search-engine/backend/internal/service"
	"search-engine/backend/pkg/cache"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "search-engine/backend/docs" // Swagger docs
)

// @title           Search Engine API
// @version         1.0
// @description     A search engine service that aggregates content from multiple providers, ranks them by relevance score, and provides search, filtering, sorting, and pagination capabilities.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  rezanicgil@gmail.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @schemes   http https

// App holds all application dependencies
type App struct {
	config        *config.Config
	router        *gin.Engine
	server        *http.Server
	redisClient   *redis.Client
	cacheInstance cache.Cache
	startTime     time.Time // Track server start time for uptime calculation
}

func main() {
	// Load and validate configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}

	// Set Gin mode
	setupGinMode()

	// Initialize database
	if err := initializeDatabase(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repository.Close()

	// Create application instance
	app := &App{
		config:    cfg,
		router:    gin.New(),
		startTime: time.Now(),
	}

	// Initialize cache and Redis
	if err := app.initializeCache(); err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}

	// Setup middleware
	app.setupMiddleware()

	// Setup routes
	app.setupRoutes()

	// Create HTTP server
	app.createServer()

	// Start initial sync from providers in background
	// This ensures data is available when the server starts
	go app.syncProvidersOnStartup(cfg)

	// Start server with graceful shutdown
	app.startServerWithGracefulShutdown()
}

// setupGinMode configures Gin framework mode based on environment
func setupGinMode() {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
}

// initializeDatabase connects to database and runs migrations
func initializeDatabase(cfg *config.Config) error {
	if err := repository.Connect(cfg); err != nil {
		return err
	}

	// Run database migrations
	// Try multiple paths to support both local and Docker environments
	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = filepath.Join("..", "..", "migrations")
	}
	migrator := migration.NewMigrator(repository.GetDB(), migrationsDir)
	if err := migrator.Run(); err != nil {
		return err
	}

	log.Println("Database initialized and migrations completed")
	return nil
}

// initializeCache initializes cache (Redis or in-memory fallback)
func (a *App) initializeCache() error {
	if !a.config.Redis.Enabled {
		log.Println("Using in-memory cache")
		cacheTTL := time.Duration(a.config.Search.CacheTTLSeconds) * time.Second
		a.cacheInstance = cache.NewInMemoryCache(cacheTTL)
		return nil
	}

	// Initialize Redis client
	a.redisClient = redis.NewClient(&redis.Options{
		Addr:     a.config.Redis.Addr,
		Password: a.config.Redis.Password,
		DB:       a.config.Redis.DB,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed, falling back to in-memory cache: %v", err)
		cacheTTL := time.Duration(a.config.Search.CacheTTLSeconds) * time.Second
		a.cacheInstance = cache.NewInMemoryCache(cacheTTL)
		a.redisClient = nil
		return nil
	}

	log.Println("Redis cache connected successfully")
	a.cacheInstance = &cache.RedisCacheWrapper{Client: a.redisClient}
	return nil
}

// setupMiddleware configures all middleware for the router
func (a *App) setupMiddleware() {
	// Global middleware
	a.router.Use(middleware.LoggerMiddleware())
	a.router.Use(middleware.CORSMiddleware())
	a.router.Use(middleware.SecurityHeadersMiddleware())

	// Error handling middleware (should be early in the chain)
	a.router.Use(middleware.ErrorHandlerMiddleware())

	// Rate limiting middleware
	rateLimiter := a.createRateLimiter()
	a.router.Use(rateLimiter)
}

// createRateLimiter creates appropriate rate limiter (Redis or in-memory)
func (a *App) createRateLimiter() gin.HandlerFunc {
	if a.config.Redis.Enabled && a.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.redisClient.Ping(ctx).Err(); err == nil {
			log.Println("Using Redis-based rate limiting")
			return middleware.NewRedisRateLimiterMiddleware(middleware.RedisRateLimiterConfig{
				Client:            a.redisClient,
				RequestsPerMinute: a.config.Rate.RequestsPerMinute,
				KeyPrefix:         "ratelimit:",
			})
		}
		log.Println("Using in-memory rate limiting (Redis unavailable)")
	} else {
		log.Println("Using in-memory rate limiting")
	}

	return middleware.NewIPRateLimiterMiddleware(middleware.RateLimiterConfig{
		RequestsPerMinute: a.config.Rate.RequestsPerMinute,
	})
}

// setupRoutes configures all API routes
func (a *App) setupRoutes() {
	// Health check endpoint (before rate limiting)
	a.router.GET("/health", a.healthCheck)

	// API v1 routes
	api := a.router.Group("/api/v1")
	a.setupAPIRoutes(api)

	// Swagger documentation
	a.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// setupAPIRoutes configures API v1 endpoints
func (a *App) setupAPIRoutes(api *gin.RouterGroup) {
	// Initialize repositories
	contentRepo := repository.NewContentRepository(repository.GetDB(), a.config.Search.MinFullTextLength)
	providerRepo := repository.NewProviderRepository(repository.GetDB())

	// Initialize services
	cacheTTL := time.Duration(a.config.Search.CacheTTLSeconds) * time.Second
	queryTimeout := time.Duration(a.config.Search.QueryTimeoutSeconds) * time.Second
	simpleQueryTimeout := time.Duration(a.config.Search.SimpleQueryTimeoutSeconds) * time.Second
	searchService := service.NewSearchService(contentRepo, a.cacheInstance, cacheTTL, queryTimeout, simpleQueryTimeout)

	// Initialize handlers
	searchHandler := handler.NewSearchHandler(searchService)
	contentHandler := handler.NewContentHandler(contentRepo, simpleQueryTimeout)
	providerHandler := handler.NewProviderHandler(providerRepo)
	statsHandler := handler.NewStatsHandler(contentRepo, providerRepo)

	// Search endpoints
	api.GET("/search", searchHandler.Search)

	// Content endpoints
	api.GET("/content/:id", contentHandler.GetContentByID)

	// Provider endpoints
	api.GET("/providers", providerHandler.GetProviders)

	// Statistics endpoints
	api.GET("/stats", statsHandler.GetStats)
}

// healthCheck handles health check requests
// Returns detailed system status including database and Redis connectivity
//
// @Summary     Health check
// @Description Get detailed system health status including database and Redis connectivity, uptime, and component statistics
// @Tags        health
// @Accept      json
// @Produce     json
// @Success     200  {object}  map[string]interface{}  "System is healthy"
// @Success     503  {object}  map[string]interface{}  "System is degraded (some components unhealthy)"
// @Router      /health [get]
func (a *App) healthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	health := gin.H{
		"status":     "OK",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"uptime":     time.Since(a.startTime).String(),
		"version":    "1.0.0",
		"components": gin.H{},
	}

	// Check database connectivity
	dbStatus := gin.H{
		"status": "healthy",
		"type":   "MySQL",
	}
	if err := repository.GetDB().PingContext(ctx); err != nil {
		dbStatus["status"] = "unhealthy"
		dbStatus["error"] = err.Error()
		health["status"] = "degraded"
	} else {
		// Get database stats
		stats := repository.GetDB().Stats()
		dbStatus["stats"] = gin.H{
			"open_connections":     stats.OpenConnections,
			"in_use":               stats.InUse,
			"idle":                 stats.Idle,
			"wait_count":           stats.WaitCount,
			"wait_duration":        stats.WaitDuration.String(),
			"max_idle_closed":      stats.MaxIdleClosed,
			"max_idle_time_closed": stats.MaxIdleTimeClosed,
			"max_lifetime_closed":  stats.MaxLifetimeClosed,
		}
	}
	health["components"].(gin.H)["database"] = dbStatus

	// Check Redis connectivity
	redisStatus := gin.H{
		"status": "not_configured",
		"type":   "Redis",
	}
	if a.config.Redis.Enabled {
		if a.redisClient != nil {
			redisCtx, redisCancel := context.WithTimeout(ctx, 2*time.Second)
			if err := a.redisClient.Ping(redisCtx).Err(); err != nil {
				redisStatus["status"] = "unhealthy"
				redisStatus["error"] = err.Error()
				health["status"] = "degraded"
			} else {
				redisStatus["status"] = "healthy"
				// Get Redis info
				info, err := a.redisClient.Info(redisCtx, "server").Result()
				if err == nil {
					redisStatus["info_available"] = true
					// Extract version if available
					if len(info) > 0 {
						redisStatus["info_length"] = len(info)
					}
				}
			}
			redisCancel()
		} else {
			redisStatus["status"] = "unavailable"
			redisStatus["message"] = "Redis client not initialized, using in-memory cache"
		}
	} else {
		redisStatus["status"] = "disabled"
		redisStatus["message"] = "Redis is disabled, using in-memory cache"
	}
	health["components"].(gin.H)["redis"] = redisStatus

	// Determine overall status code
	statusCode := http.StatusOK
	if health["status"] == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"health": health,
	})
}

// createServer creates and configures the HTTP server
func (a *App) createServer() {
	a.server = &http.Server{
		Addr:         a.config.Server.Host + ":" + a.config.Server.Port,
		Handler:      a.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// startServerWithGracefulShutdown starts the server and handles graceful shutdown
func (a *App) startServerWithGracefulShutdown() {
	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s:%s", a.config.Server.Host, a.config.Server.Port)
		log.Printf("API documentation available at http://%s:%s/swagger/index.html", a.config.Server.Host, a.config.Server.Port)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// syncProvidersOnStartup syncs data from providers when the server starts
func (a *App) syncProvidersOnStartup(cfg *config.Config) {
	if os.Getenv("AUTO_SYNC_ON_START") == "false" {
		return
	}

	time.Sleep(2 * time.Second)
	log.Println("Starting initial provider sync...")

	providerRepo := repository.NewProviderRepository(repository.GetDB())
	contentRepo := repository.NewContentRepository(repository.GetDB(), cfg.Search.MinFullTextLength)
	tagRepo := repository.NewContentTagRepository(repository.GetDB())
	manager := provider.NewManager(providerRepo, contentRepo, tagRepo)

	// Ensure providers exist and register them
	providers := []struct {
		name, url string
		format    model.ProviderFormat
	}{
		{"provider1", cfg.Provider.Provider1URL, model.ProviderFormatJSON},
		{"provider2", cfg.Provider.Provider2URL, model.ProviderFormatXML},
	}

	for _, p := range providers {
		existing, err := providerRepo.GetByName(p.name)
		if err != nil && errors.Is(err, repository.ErrProviderNotFound) {
			if err := providerRepo.Create(&model.Provider{
				Name:               p.name,
				URL:                p.url,
				Format:             p.format,
				RateLimitPerMinute: 60,
			}); err != nil {
				log.Printf("Warning: Failed to create provider %s: %v", p.name, err)
			}
		} else if err == nil {
			existing.URL = p.url
			existing.Format = p.format
			existing.RateLimitPerMinute = 60
			providerRepo.Update(existing)
		}

		if p.format == model.ProviderFormatJSON {
			manager.RegisterProvider(provider.NewJSONProvider(p.name, p.url))
		} else {
			manager.RegisterProvider(provider.NewXMLProvider(p.name, p.url))
		}
	}

	if err := manager.FetchAll(); err != nil {
		log.Printf("Warning: Failed to fetch from providers: %v", err)
		return
	}

	// Recalculate scores
	scoringService := service.NewScoringService(contentRepo)
	allProviders, _ := providerRepo.GetAll()
	for _, p := range allProviders {
		scoringService.RecalculateScoresForProvider(p.ID)
	}

	log.Println("Initial provider sync completed")
}
