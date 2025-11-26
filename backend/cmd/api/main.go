// main.go - Main entry point for the Search Engine API server
// Initializes database, services, handlers, and starts the HTTP server
package main

import (
	"context"
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
		config: cfg,
		router: gin.New(),
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
	searchService := service.NewSearchService(contentRepo, a.cacheInstance, cacheTTL)

	// Initialize handlers
	searchHandler := handler.NewSearchHandler(searchService)
	contentHandler := handler.NewContentHandler(contentRepo)
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
func (a *App) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Search engine API is running",
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
