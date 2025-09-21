// services/nlp-processing/main.go
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

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/database"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/handler"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/repository"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/service"
)

func main() {
	// Initialize configuration
	if err := initConfig(); err != nil {
		logrus.Fatalf("Failed to initialize config: %v", err)
	}

	// Initialize logger
	initLogger()

	// Initialize database connection
	// Note: For high-volume financial analytics in production, consider ClickHouse on AWS
	// This PostgreSQL implementation provides compatibility with the requested architecture
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

	// Create additional indexes for analytics performance
	if err := db.CreateIndexes(); err != nil {
		logrus.Warnf("Failed to create some indexes: %v", err)
	}

	// Create materialized views for analytics
	if err := db.CreateMaterializedViews(); err != nil {
		logrus.Warnf("Failed to create materialized views: %v", err)
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
	analysisRepo := repository.NewAnalysisRepository(db.DB)
	cacheRepo := repository.NewCacheRepository(redisClient)

	// Initialize ML services
	sentimentService := service.NewSentimentService()
	nerService := service.NewNERService()
	topicService := service.NewTopicClassificationService()

	// Initialize main NLP service
	nlpService := service.NewNLPService(
		analysisRepo,
		cacheRepo,
		sentimentService,
		nerService,
		topicService,
	)

	// Initialize HTTP handlers
	httpHandler := handler.NewHTTPHandler(nlpService, analysisRepo)

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
			"service":   "nlp-processing",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Analysis endpoints
		v1.POST("/analyze", httpHandler.AnalyzeArticle)
		v1.POST("/analyze/batch", httpHandler.AnalyzeBatch)
		v1.GET("/analysis/:article_id", httpHandler.GetAnalysis)

		// Analytics endpoints
		v1.GET("/analytics/sentiment/:symbol", httpHandler.GetSentimentTrends)
		v1.GET("/analytics/topics/trending", httpHandler.GetTrendingTopics)
		v1.GET("/analytics/entities/mentioned", httpHandler.GetMostMentionedEntities)
		v1.GET("/analytics/breaking-news", httpHandler.GetBreakingNews)

		// Model endpoints
		v1.POST("/models/sentiment", httpHandler.PredictSentiment)
		v1.POST("/models/entities", httpHandler.ExtractEntities)
		v1.POST("/models/topics", httpHandler.ClassifyTopic)
	}

	// Start background workers
	go startBackgroundWorkers(ctx, nlpService, db)

	// Setup HTTP server
	port := viper.GetInt("server.port")
	if port == 0 {
		port = 4002 // Default port for NLP processing service
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logrus.Infof("Starting NLP Processing Service on port %d", port)
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
	viper.SetDefault("server.port", 4002)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.database", 1)

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

func startBackgroundWorkers(ctx context.Context, nlpService service.NLPService, db *database.Database) {
	logrus.Info("Starting background processing workers")

	// Start materialized view refresh worker
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := db.RefreshMaterializedViews(); err != nil {
					logrus.WithError(err).Error("Failed to refresh materialized views")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Start batch processing worker for pending articles
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := nlpService.ProcessPendingArticles(ctx); err != nil {
					logrus.WithError(err).Error("Batch processing failed")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Start model cache warming worker
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := nlpService.WarmModelCaches(ctx); err != nil {
					logrus.WithError(err).Error("Model cache warming failed")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	logrus.Info("Background workers started successfully")
}
