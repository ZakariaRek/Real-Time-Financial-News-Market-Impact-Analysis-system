// services/news-ingestion/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/database"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/handler"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/repository"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/service"
)

func main() {
	// Initialize configuration
	if err := initConfig(); err != nil {
		logrus.Fatalf("Failed to initialize config: %v", err)
	}

	// Initialize logger
	initLogger()

	// Initialize database connection
	// For AWS RDS in production, ensure proper security groups and VPC configuration
	dbConfig := database.Config{
		Host:            viper.GetString("database.postgres.host"),
		Port:            viper.GetInt("database.postgres.port"),
		Database:        viper.GetString("database.postgres.database"),
		Username:        viper.GetString("database.postgres.username"),
		Password:        viper.GetString("database.postgres.password"),
		SSLMode:         viper.GetString("database.postgres.ssl_mode"),
		MaxOpenConns:    viper.GetInt("database.postgres.max_open_conns"),
		MaxIdleConns:    viper.GetInt("database.postgres.max_idle_conns"),
		ConnMaxLifetime: viper.GetString("database.postgres.conn_max_lifetime"),
		ConnMaxIdleTime: viper.GetString("database.postgres.conn_max_idle_time"),
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.AutoMigrate(); err != nil {
		logrus.Fatalf("Failed to run migrations: %v", err)
	}

	// Create additional indexes for performance
	if err := db.CreateIndexes(); err != nil {
		logrus.Warnf("Failed to create some indexes: %v", err)
	}

	// Initialize Redis client for caching
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.database"),
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logrus.Warnf("Failed to connect to Redis: %v", err)
	} else {
		logrus.Info("Successfully connected to Redis")
	}

	// Initialize repositories
	articleRepo := repository.NewArticleRepository(db.DB)
	sourceRepo := repository.NewSourceRepository(db.DB)
	processingLogRepo := repository.NewProcessingLogRepository(db.DB)
	rateLimitRepo := repository.NewRateLimitRepository(db.DB)

	// Initialize cache repository
	cacheRepo := repository.NewCacheRepository(redisClient)

	// Initialize services
	deduplicationService := service.NewDeduplicationService(articleRepo, cacheRepo)
	scraperService := service.NewScraperService(sourceRepo, rateLimitRepo)
	ingestionService := service.NewIngestionService(
		articleRepo,
		sourceRepo,
		processingLogRepo,
		deduplicationService,
		scraperService,
	)

	// Initialize HTTP handlers
	httpHandler := handler.NewHTTPHandler(ingestionService, articleRepo, sourceRepo)

	// Setup Gin router
	if viper.GetString("server.environment") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "news-ingestion",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Article endpoints
		v1.POST("/articles", httpHandler.CreateArticle)
		v1.GET("/articles/:id", httpHandler.GetArticle)
		v1.GET("/articles", httpHandler.ListArticles)
		v1.PUT("/articles/:id/status", httpHandler.UpdateArticleStatus)

		// Source endpoints
		v1.POST("/sources", httpHandler.CreateSource)
		v1.GET("/sources", httpHandler.ListSources)
		v1.GET("/sources/:id", httpHandler.GetSource)
		v1.PUT("/sources/:id", httpHandler.UpdateSource)

		// Ingestion endpoints
		v1.POST("/ingest/manual", httpHandler.TriggerManualIngestion)
		v1.GET("/ingest/status", httpHandler.GetIngestionStatus)
	}

	// Start background ingestion workers
	go startBackgroundWorkers(ctx, ingestionService)

	// Setup HTTP server
	port := viper.GetInt("server.port")
	if port == 0 {
		port = 4001 // Default port for news ingestion service
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logrus.Infof("Starting News Ingestion Service on port %d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Set default values
	viper.SetDefault("server.port", 4001)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)

	// Read environment variables
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logrus.Warnf("Config file not found, using defaults: %v", err)
	}

	return nil
}

func initLogger() {
	level := viper.GetString("logging.level")
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	format := viper.GetString("logging.format")
	if format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
	}
}

func startBackgroundWorkers(ctx context.Context, ingestionService service.IngestionService) {
	logrus.Info("Starting background ingestion workers")

	// Start RSS ingestion worker
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := ingestionService.IngestFromRSS(ctx); err != nil {
					logrus.WithError(err).Error("RSS ingestion failed")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Start News API ingestion worker
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := ingestionService.IngestFromNewsAPI(ctx); err != nil {
					logrus.WithError(err).Error("News API ingestion failed")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Start rate limit cleanup worker
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := ingestionService.CleanupRateLimitTracking(ctx); err != nil {
					logrus.WithError(err).Error("Rate limit cleanup failed")
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
