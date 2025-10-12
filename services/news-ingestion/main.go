// services/news-ingestion/main.go
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/client"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/database"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/handler"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/repository"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/service"
	newsv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/proto/services/news-ingestion/proto/gen"
)

// ServiceContainer holds all initialized services
type ServiceContainer struct {
	articleRepo             repository.ArticleRepository
	sourceRepo              repository.SourceRepository
	logRepo                 repository.ProcessingLogRepository
	rateLimitRepo           repository.RateLimitRepository
	deduplicationService    service.DeduplicationService
	extractionService       *service.DataExtractionService
	ingestionService        service.IngestionService
	sentimentTriggerService *service.SentimentTriggerService
	httpHandler             *handler.HTTPHandler
	grpcHandler             *handler.GRPCHandler
	newsAPIClient           client.NewsAPIClient
	rssClient               client.RSSClient
	twitterClient           client.TwitterClient
	nlpClient               client.NLPProcessingClient
	db                      *database.Database
	cronScheduler           *cron.Cron
}

type Metrics struct {
	ArticlesIngested  int64
	ArticlesProcessed int64
	Errors            int64
	LastIngestion     time.Time
	mu                sync.RWMutex
}

var (
	serviceMetrics = &Metrics{}
	startTime      = time.Now()
)

func main() {
	// Initialize configuration first
	if err := initConfig(); err != nil {
		logrus.Fatalf("Failed to initialize config: %v", err)
	}

	// Initialize logger based on config
	initLogger()

	logrus.Info("========================================")
	logrus.Info("   News Ingestion Service Starting")
	logrus.Info("========================================")
	logrus.Infof("Environment: %s", viper.GetString("SERVER_ENVIRONMENT"))
	if viper.ConfigFileUsed() != "" {
		logrus.Infof("Config loaded from: %s", viper.ConfigFileUsed())
	}

	// Initialize service container
	container, err := initializeServices()
	if err != nil {
		logrus.Fatalf("Failed to initialize services: %v", err)
	}
	defer container.cleanup()

	// Get ports from config
	httpPort := viper.GetInt("SERVER_HTTP_PORT")
	grpcPort := viper.GetInt("SERVER_GRPC_PORT")

	httpServer := setupHTTPServer(container.httpHandler, httpPort, container.db, container.ingestionService)
	grpcServer := setupGRPCServer(container.grpcHandler, grpcPort)

	// Start background jobs if enabled
	if viper.GetBool("ENABLE_AUTO_INGESTION") {
		logrus.Info("Starting background ingestion jobs...")
		container.startBackgroundJobs()
	} else {
		logrus.Info("Auto-ingestion disabled. Use manual triggers.")
	}

	// Start servers
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		startHTTPServer(httpServer, httpPort)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		startGRPCServer(grpcServer, grpcPort)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		reportMetricsPeriodically(ctx)
	}()

	logrus.Info("========================================")
	logrus.Info("   All Services Running Successfully")
	logrus.Info("========================================")
	logrus.Infof("   HTTP Server:  http://localhost:%d", httpPort)
	logrus.Infof("   gRPC Server:  localhost:%d", grpcPort)
	logrus.Infof("   Health Check: http://localhost:%d/health", httpPort)
	if container.sentimentTriggerService != nil {
		logrus.Infof("   Sentiment Trigger: ENABLED (threshold: %d)", viper.GetInt("SENTIMENT_TRIGGER_THRESHOLD"))
	} else {
		logrus.Info("   Sentiment Trigger: DISABLED")
	}
	logrus.Info("========================================")
	logrus.Info("   Press Ctrl+C to shutdown")
	logrus.Info("========================================")

	gracefulShutdown(httpServer, grpcServer, container, cancel)
	wg.Wait()
	logrus.Info("All services stopped gracefully")
}

func initConfig() error {
	// Set config file name and type
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add config paths - check multiple locations
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/app/config") // For Docker container

	// Enable environment variable reading
	viper.AutomaticEnv()
	// DO NOT set a prefix - we want to match the exact ENV var names from Kubernetes
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults before reading config
	setConfigDefaults()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logrus.Warn("⚠️  Config file not found, using environment variables and defaults")
		} else {
			logrus.Warnf("⚠️  Error reading config file: %v, using environment variables and defaults", err)
		}
	} else {
		logrus.Infof("✓ Using config file: %s", viper.ConfigFileUsed())
	}

	// Validate critical configuration
	if err := validateConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Log configuration summary (without sensitive data)
	logConfigSummary()

	return nil
}

func setConfigDefaults() {
	// Server defaults - matching Kubernetes environment variables
	viper.SetDefault("SERVER_HTTP_PORT", 4001)
	viper.SetDefault("SERVER_GRPC_PORT", 4002)
	viper.SetDefault("SERVER_ENVIRONMENT", "development")

	// Database defaults - matching Kubernetes environment variables
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", 5432)
	viper.SetDefault("POSTGRES_DB", "news_ingestion")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "postgres")
	viper.SetDefault("DATABASE_SSL_MODE", "disable")
	viper.SetDefault("DATABASE_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DATABASE_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DATABASE_CONN_MAX_LIFETIME", "5m")
	viper.SetDefault("DATABASE_CONN_MAX_IDLE_TIME", "1m")

	// Redis defaults - matching Kubernetes environment variables
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DATABASE", "0")

	// Logging defaults
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "text")

	// Processing defaults
	viper.SetDefault("ENABLE_AUTO_INGESTION", true)
	viper.SetDefault("SENTIMENT_TRIGGER_THRESHOLD", 10)
	viper.SetDefault("RSS_SCHEDULE", "0 */2 * * * *")
	viper.SetDefault("NEWSAPI_SCHEDULE", "0 */7 * * * *")
	viper.SetDefault("CLEANUP_SCHEDULE", "0 0 0 * * *")
	viper.SetDefault("BATCH_SIZE", 100)
	viper.SetDefault("WORKER_COUNT", 5)
	viper.SetDefault("RETRY_ATTEMPTS", 3)
	viper.SetDefault("TIMEOUT_SECONDS", 30)

	// News sources defaults
	viper.SetDefault("NEWSAPI_KEY", "")
	viper.SetDefault("NEWSAPI_BASE_URL", "https://newsapi.org/v2")
	viper.SetDefault("NEWSAPI_ENABLED", true)
	viper.SetDefault("TWITTER_BEARER_TOKEN", "")
	viper.SetDefault("TWITTER_BASE_URL", "https://api.twitter.com/2")
	viper.SetDefault("TWITTER_ENABLED", false)

	// External services defaults
	viper.SetDefault("NLP_PROCESSING_HOST", "localhost")
	viper.SetDefault("NLP_PROCESSING_PORT", "50052")
	viper.SetDefault("NLP_TIMEOUT", "30s")
	viper.SetDefault("NLP_MAX_RETRY", 3)
}

func validateConfig() error {
	// Validate critical settings
	if viper.GetString("POSTGRES_HOST") == "" {
		return fmt.Errorf("database host is required (POSTGRES_HOST)")
	}

	if viper.GetString("POSTGRES_DB") == "" {
		return fmt.Errorf("database name is required (POSTGRES_DB)")
	}

	if viper.GetInt("SERVER_HTTP_PORT") <= 0 || viper.GetInt("SERVER_HTTP_PORT") > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", viper.GetInt("SERVER_HTTP_PORT"))
	}

	if viper.GetInt("SERVER_GRPC_PORT") <= 0 || viper.GetInt("SERVER_GRPC_PORT") > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", viper.GetInt("SERVER_GRPC_PORT"))
	}

	return nil
}

func logConfigSummary() {
	logrus.Info("Configuration Summary:")
	logrus.Infof("  HTTP Port: %d", viper.GetInt("SERVER_HTTP_PORT"))
	logrus.Infof("  gRPC Port: %d", viper.GetInt("SERVER_GRPC_PORT"))
	logrus.Infof("  Environment: %s", viper.GetString("SERVER_ENVIRONMENT"))
	logrus.Infof("  Database Host: %s", viper.GetString("POSTGRES_HOST"))
	logrus.Infof("  Database Name: %s", viper.GetString("POSTGRES_DB"))
	logrus.Infof("  Database Port: %d", viper.GetInt("POSTGRES_PORT"))
	logrus.Infof("  Redis Host: %s", viper.GetString("REDIS_HOST"))
	logrus.Infof("  Redis Port: %s", viper.GetString("REDIS_PORT"))
	logrus.Infof("  Log Level: %s", viper.GetString("LOG_LEVEL"))
	logrus.Infof("  Auto Ingestion: %v", viper.GetBool("ENABLE_AUTO_INGESTION"))
	logrus.Infof("  Sentiment Threshold: %d", viper.GetInt("SENTIMENT_TRIGGER_THRESHOLD"))

	// Check if NewsAPI key is configured
	if viper.GetString("NEWSAPI_KEY") != "" {
		logrus.Info("  NewsAPI: Configured ✓")
	} else {
		logrus.Warn("  NewsAPI: Not configured (API key missing)")
	}

	// Check if NLP service is configured
	nlpHost := viper.GetString("NLP_PROCESSING_HOST")
	nlpPort := viper.GetString("NLP_PROCESSING_PORT")
	logrus.Infof("  NLP Service: %s:%s", nlpHost, nlpPort)
}

func initLogger() {
	level := viper.GetString("LOG_LEVEL")
	switch strings.ToLower(level) {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	format := viper.GetString("LOG_FORMAT")
	if strings.ToLower(format) == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableColors:   false,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	logrus.SetOutput(os.Stdout)
	logrus.Infof("✓ Logger initialized (level: %s, format: %s)", level, format)
}

func initDatabase() (*database.Database, error) {
	dbConfig := database.Config{
		Host:            viper.GetString("POSTGRES_HOST"),
		Port:            viper.GetInt("POSTGRES_PORT"),
		Database:        viper.GetString("POSTGRES_DB"),
		Username:        viper.GetString("POSTGRES_USER"),
		Password:        viper.GetString("POSTGRES_PASSWORD"),
		SSLMode:         viper.GetString("DATABASE_SSL_MODE"),
		MaxOpenConns:    viper.GetInt("DATABASE_MAX_OPEN_CONNS"),
		MaxIdleConns:    viper.GetInt("DATABASE_MAX_IDLE_CONNS"),
		ConnMaxLifetime: viper.GetString("DATABASE_CONN_MAX_LIFETIME"),
		ConnMaxIdleTime: viper.GetString("DATABASE_CONN_MAX_IDLE_TIME"),
	}

	logrus.Infof("Connecting to database at %s:%d/%s",
		dbConfig.Host, dbConfig.Port, dbConfig.Database)

	var db *database.Database
	var err error

	// Retry connection with exponential backoff
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = database.NewConnection(dbConfig)
		if err == nil {
			return db, nil
		}

		if i < maxRetries-1 {
			waitTime := time.Duration(i+1) * 2 * time.Second
			logrus.Warnf("Failed to connect to database (attempt %d/%d): %v. Retrying in %v...",
				i+1, maxRetries, err, waitTime)
			time.Sleep(waitTime)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

func initializeServices() (*ServiceContainer, error) {
	logrus.Info("Initializing services...")

	db, err := initDatabase()
	if err != nil {
		return nil, fmt.Errorf("database initialization failed: %w", err)
	}
	logrus.Info("✓ Database connection established")

	if err := db.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	logrus.Info("✓ Database migrations completed")

	if err := db.SeedData(context.Background()); err != nil {
		logrus.WithError(err).Warn("Failed to seed database, continuing anyway")
	} else {
		logrus.Info("✓ Database seeding completed")
	}

	if err := db.CreateIndexes(); err != nil {
		logrus.Warnf("Some indexes failed to create: %v", err)
	} else {
		logrus.Info("✓ Database indexes created")
	}

	articleRepo := repository.NewArticleRepository(db.DB)
	sourceRepo := repository.NewSourceRepository(db.DB)
	logRepo := repository.NewProcessingLogRepository(db.DB)
	rateLimitRepo := repository.NewRateLimitRepository(db.DB)
	logrus.Info("✓ Repositories initialized")

	newsAPIClient := client.NewNewsAPIClient(
		viper.GetString("NEWSAPI_KEY"),
		viper.GetString("NEWSAPI_BASE_URL"),
	)
	rssClient := client.NewRSSClient()
	twitterClient := client.NewTwitterClient(
		viper.GetString("TWITTER_BEARER_TOKEN"),
		viper.GetString("TWITTER_BASE_URL"),
	)
	logrus.Info("✓ External API clients initialized")

	nlpHost := viper.GetString("NLP_PROCESSING_HOST")
	nlpPort := viper.GetString("NLP_PROCESSING_PORT")
	nlpEndpoint := fmt.Sprintf("%s:%s", nlpHost, nlpPort)

	nlpClient, err := client.NewNLPProcessingClient(nlpEndpoint)
	if err != nil {
		logrus.WithError(err).Warn("⚠️ Failed to connect to NLP service, sentiment trigger will be disabled")
		nlpClient = nil
	} else {
		logrus.Info("✓ NLP client initialized")
	}

	deduplicationService := service.NewDeduplicationService(articleRepo)
	extractionService := service.NewDataExtractionService(
		articleRepo,
		sourceRepo,
		deduplicationService,
	)
	ingestionService := service.NewIngestionService(
		articleRepo,
		sourceRepo,
		logRepo,
		rateLimitRepo,
		deduplicationService,
		newsAPIClient,
		rssClient,
		twitterClient,
	)
	logrus.Info("✓ Core services initialized")

	var sentimentTriggerService *service.SentimentTriggerService
	if nlpClient != nil {
		threshold := viper.GetInt("SENTIMENT_TRIGGER_THRESHOLD")
		sentimentTriggerService = service.NewSentimentTriggerService(
			articleRepo,
			nlpClient,
			threshold,
		)
		logrus.Infof("✓ Sentiment trigger service initialized (threshold: %d articles)", threshold)
	}

	httpHandler := handler.NewHTTPHandler(ingestionService, articleRepo, sourceRepo)
	grpcHandler := handler.NewGRPCHandler(ingestionService, articleRepo, sourceRepo, logRepo)
	logrus.Info("✓ Handlers initialized")

	cronScheduler := cron.New(cron.WithSeconds())
	logrus.Info("✓ Cron scheduler initialized")

	return &ServiceContainer{
		articleRepo:             articleRepo,
		sourceRepo:              sourceRepo,
		logRepo:                 logRepo,
		rateLimitRepo:           rateLimitRepo,
		deduplicationService:    deduplicationService,
		extractionService:       extractionService,
		ingestionService:        ingestionService,
		sentimentTriggerService: sentimentTriggerService,
		httpHandler:             httpHandler,
		grpcHandler:             grpcHandler,
		newsAPIClient:           newsAPIClient,
		rssClient:               rssClient,
		twitterClient:           twitterClient,
		nlpClient:               nlpClient,
		db:                      db,
		cronScheduler:           cronScheduler,
	}, nil
}

func (c *ServiceContainer) startBackgroundJobs() {
	ctx := context.Background()

	// RSS Feed Ingestion
	rssSchedule := viper.GetString("RSS_SCHEDULE")
	_, err := c.cronScheduler.AddFunc(rssSchedule, func() {
		logrus.Info("🔄 Starting scheduled RSS ingestion...")
		startTime := time.Now()

		if err := c.ingestionService.IngestFromRSS(ctx); err != nil {
			logrus.WithError(err).Error("❌ Scheduled RSS ingestion failed")
			incrementErrorMetric()
		} else {
			duration := time.Since(startTime)
			logrus.Infof("✅ RSS ingestion completed in %v", duration)
			updateIngestionMetrics()
		}
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to schedule RSS ingestion")
	} else {
		logrus.Infof("✓ Scheduled RSS ingestion: %s", rssSchedule)
	}

	// NewsAPI Ingestion
	newsAPISchedule := viper.GetString("NEWSAPI_SCHEDULE")
	_, err = c.cronScheduler.AddFunc(newsAPISchedule, func() {
		logrus.Info("🔄 Starting scheduled NewsAPI ingestion...")
		startTime := time.Now()

		if err := c.ingestionService.IngestFromNewsAPI(ctx); err != nil {
			logrus.WithError(err).Error("❌ Scheduled NewsAPI ingestion failed")
			incrementErrorMetric()
		} else {
			duration := time.Since(startTime)
			logrus.Infof("✅ NewsAPI ingestion completed in %v", duration)
			updateIngestionMetrics()
		}
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to schedule NewsAPI ingestion")
	} else {
		logrus.Infof("✓ Scheduled NewsAPI ingestion: %s", newsAPISchedule)
	}

	// Rate limit cleanup
	cleanupSchedule := viper.GetString("CLEANUP_SCHEDULE")
	_, err = c.cronScheduler.AddFunc(cleanupSchedule, func() {
		logrus.Info("🧹 Starting rate limit cleanup...")
		cutoff := time.Now().AddDate(0, 0, -7)

		if err := c.rateLimitRepo.CleanupOldRecords(ctx, cutoff); err != nil {
			logrus.WithError(err).Error("❌ Rate limit cleanup failed")
		} else {
			logrus.Info("✅ Rate limit cleanup completed")
		}
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to schedule rate limit cleanup")
	} else {
		logrus.Infof("✓ Scheduled rate limit cleanup: %s", cleanupSchedule)
	}

	c.cronScheduler.Start()
	logrus.Info("✓ Background jobs started successfully")

	// Start sentiment trigger service
	if c.sentimentTriggerService != nil {
		go func() {
			logrus.Info("🚀 Starting sentiment trigger service...")
			c.sentimentTriggerService.Start(ctx)
		}()
		logrus.Info("✓ Sentiment trigger service started")
	}
}

func (c *ServiceContainer) cleanup() {
	logrus.Info("🧹 Cleaning up resources...")

	if c.cronScheduler != nil {
		c.cronScheduler.Stop()
		logrus.Info("✓ Stopped cron scheduler")
	}

	if c.nlpClient != nil {
		if err := c.nlpClient.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close NLP client")
		} else {
			logrus.Info("✓ Closed NLP client connection")
		}
	}

	if c.db != nil {
		if err := c.db.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close database connection")
		} else {
			logrus.Info("✓ Closed database connection")
		}
	}
}

func setupHTTPServer(httpHandler *handler.HTTPHandler, port int, db *database.Database, ingestionService service.IngestionService) *http.Server {
	if viper.GetString("SERVER_ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "news-ingestion",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC(),
			"database":  getDatabaseStatus(db),
			"servers": gin.H{
				"http": "running",
				"grpc": "running",
			},
			"metrics": getMetrics(),
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		if getDatabaseStatus(db) == "connected" {
			c.JSON(http.StatusOK, gin.H{"status": "ready"})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
		}
	})

	router.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	router.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, getMetrics())
	})

	router.GET("/db/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    getDatabaseStatus(db),
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Article routes
		articles := v1.Group("/articles")
		{
			articles.POST("", httpHandler.CreateArticle)
			articles.GET("/:id", httpHandler.GetArticle)
			articles.GET("", httpHandler.ListArticles)
			articles.PUT("/:id/status", httpHandler.UpdateArticleStatus)
		}

		// Source routes
		sources := v1.Group("/sources")
		{
			sources.POST("", httpHandler.CreateSource)
			sources.GET("/:id", httpHandler.GetSource)
			sources.GET("", httpHandler.ListSources)
			sources.PUT("/:id", httpHandler.UpdateSource)
		}

		// Ingestion routes
		ingestion := v1.Group("/ingestion")
		{
			ingestion.POST("/trigger", httpHandler.TriggerManualIngestion)
			ingestion.GET("/status", httpHandler.GetIngestionStatus)

			ingestion.POST("/trigger/rss", func(c *gin.Context) {
				logrus.Info("📰 Manual RSS ingestion triggered via HTTP")
				ctx := context.Background()

				if err := ingestionService.IngestFromRSS(ctx); err != nil {
					logrus.WithError(err).Error("❌ Manual RSS ingestion failed")
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   err.Error(),
						"message": "RSS ingestion failed",
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message":   "✅ RSS ingestion completed successfully",
					"timestamp": time.Now().UTC(),
				})
			})

			ingestion.POST("/trigger/newsapi", func(c *gin.Context) {
				logrus.Info("📡 Manual NewsAPI ingestion triggered via HTTP")
				ctx := context.Background()

				if err := ingestionService.IngestFromNewsAPI(ctx); err != nil {
					logrus.WithError(err).Error("❌ Manual NewsAPI ingestion failed")
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   err.Error(),
						"message": "NewsAPI ingestion failed",
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message":   "✅ NewsAPI ingestion completed successfully",
					"timestamp": time.Now().UTC(),
				})
			})
		}
	}

	return &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func setupGRPCServer(grpcHandler *handler.GRPCHandler, port int) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 4),
		grpc.MaxSendMsgSize(1024 * 1024 * 4),
		grpc.ConnectionTimeout(30 * time.Second),
	}

	grpcServer := grpc.NewServer(opts...)
	newsv1.RegisterNewsServiceServer(grpcServer, grpcHandler)

	if viper.GetString("SERVER_ENVIRONMENT") != "production" {
		reflection.Register(grpcServer)
	}

	return grpcServer
}

func startHTTPServer(server *http.Server, port int) {
	logrus.Infof("🌐 Starting HTTP server on port %d", port)
	logrus.Infof("  → Health:  http://localhost:%d/health", port)
	logrus.Infof("  → Ready:   http://localhost:%d/ready", port)
	logrus.Infof("  → Metrics: http://localhost:%d/metrics", port)
	logrus.Infof("  → API:     http://localhost:%d/api/v1", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("HTTP server failed: %v", err)
	}
}

func startGRPCServer(server *grpc.Server, port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logrus.Fatalf("Failed to listen on gRPC port %d: %v", port, err)
	}

	logrus.Infof("🔌 Starting gRPC server on port %d", port)
	logrus.Infof("  → Address: localhost:%d", port)
	logrus.Infof("  → Reflection: enabled")

	if err := server.Serve(lis); err != nil {
		logrus.Fatalf("gRPC server failed: %v", err)
	}
}

func gracefulShutdown(httpServer *http.Server, grpcServer *grpc.Server, container *ServiceContainer, cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("========================================")
	logrus.Info("   Shutdown signal received")
	logrus.Info("========================================")

	ctx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	cancel()

	logrus.Info("Shutting down HTTP server...")
	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Errorf("HTTP server shutdown error: %v", err)
	} else {
		logrus.Info("✓ HTTP server stopped")
	}

	logrus.Info("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	logrus.Info("✓ gRPC server stopped")

	container.cleanup()
}

func reportMetricsPeriodically(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := getMetrics()
			logrus.WithFields(logrus.Fields{
				"articles_ingested":  metrics["articles_ingested"],
				"articles_processed": metrics["articles_processed"],
				"errors":             metrics["errors"],
				"last_ingestion":     metrics["last_ingestion"],
			}).Info("📊 Service metrics report")
		}
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func getDatabaseStatus(db *database.Database) string {
	if db == nil {
		return "disconnected"
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		return "error"
	}

	if err := sqlDB.Ping(); err != nil {
		return "unhealthy"
	}

	return "connected"
}

func getMetrics() map[string]interface{} {
	serviceMetrics.mu.RLock()
	defer serviceMetrics.mu.RUnlock()

	return map[string]interface{}{
		"articles_ingested":  serviceMetrics.ArticlesIngested,
		"articles_processed": serviceMetrics.ArticlesProcessed,
		"errors":             serviceMetrics.Errors,
		"last_ingestion":     serviceMetrics.LastIngestion.Format(time.RFC3339),
		"uptime":             time.Since(startTime).String(),
	}
}

func updateIngestionMetrics() {
	serviceMetrics.mu.Lock()
	defer serviceMetrics.mu.Unlock()
	serviceMetrics.ArticlesIngested++
	serviceMetrics.LastIngestion = time.Now()
}

func incrementErrorMetric() {
	serviceMetrics.mu.Lock()
	defer serviceMetrics.mu.Unlock()
	serviceMetrics.Errors++
}
