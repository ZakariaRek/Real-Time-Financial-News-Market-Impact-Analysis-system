// services/nlp-processing/main.go
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	//newsv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/proto/services/news-ingestion/proto/gen"
	newsv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/proto/gen/news/v1"
	nlpv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/proto/gen/nlp/v1"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

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

	logrus.Info("Starting NLP Processing Service...")

	// Initialize database connection
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

	// Connect to database
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logrus.Info("Database connection established successfully")

	// Run database migrations
	if err := db.AutoMigrate(); err != nil {
		logrus.Fatalf("Failed to run migrations: %v", err)
	}

	logrus.Info("Database migrations completed successfully")

	// Create additional indexes for performance
	if err := db.CreateIndexes(); err != nil {
		logrus.Warnf("Failed to create some indexes: %v", err)
	}

	// Create materialized views for analytics
	if err := db.CreateMaterializedViews(); err != nil {
		logrus.Warnf("Failed to create materialized views: %v", err)
	}

	// Initialize Redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.database"),
	})
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logrus.Warnf("Redis connection failed: %v", err)
	} else {
		logrus.Info("Redis connection established successfully")
	}

	// Initialize repositories
	articleRepo := repository.NewArticleRepository(db.DB)
	sourceRepo := repository.NewSourceRepository(db.DB)
	logRepo := repository.NewProcessingLogRepository(db.DB)
	analysisRepo := repository.NewAnalysisRepository(db.DB)
	cacheRepo := repository.NewCacheRepository(redisClient)

	// Initialize NLP services
	sentimentConfig := service.SentimentConfig{
		ModelPath: viper.GetString("models.finbert.model_path"),
		Device:    viper.GetString("models.finbert.device"),
		BatchSize: viper.GetInt("models.finbert.batch_size"),
		MaxLength: viper.GetInt("models.finbert.max_length"),
	}
	sentimentService := service.NewSentimentService(sentimentConfig)

	nerConfig := service.NERConfig{
		ModelPath:           viper.GetString("models.ner.model_path"),
		SpacyModel:          viper.GetString("models.ner.spacy_model"),
		ConfidenceThreshold: float32(viper.GetFloat64("models.ner.confidence_threshold")),
	}
	nerService := service.NewNERService(nerConfig)

	topicConfig := service.TopicConfig{
		ModelPath:  viper.GetString("models.topic_classifier.model_path"),
		Categories: viper.GetStringSlice("models.topic_classifier.categories"),
	}
	topicService := service.NewTopicService(topicConfig)

	// Initialize main NLP processing service
	nlpConfig := service.NLPConfig{
		WorkerCount:    viper.GetInt("processing.worker_count"),
		BatchSize:      viper.GetInt("processing.batch_size"),
		TimeoutSeconds: viper.GetInt("processing.timeout_seconds"),
		RetryAttempts:  viper.GetInt("processing.retry_attempts"),
	}
	nlpService := service.NewNLPProcessingService(
		sentimentService,
		nerService,
		topicService,
		analysisRepo,
		cacheRepo,
		nlpConfig,
	)

	// Initialize models
	logrus.Info("Initializing ML models...")
	if err := nlpService.InitializeModels(context.Background()); err != nil {
		logrus.Warnf("Failed to initialize some models: %v", err)
	} else {
		logrus.Info("All ML models initialized successfully")
	}

	// Initialize ingestion service (placeholder)
	ingestionService := service.NewIngestionService()

	// Initialize handlers
	httpHandler := handler.NewHTTPHandler(ingestionService, articleRepo, sourceRepo)
	grpcHandler := handler.NewGRPCHandler(ingestionService, articleRepo, sourceRepo, logRepo)
	nlpGRPCHandler := handler.NewNLPGRPCHandler(nlpService, analysisRepo)

	// Setup HTTP server
	httpPort := viper.GetInt("server.port")
	if httpPort == 0 {
		httpPort = 4002 // Default HTTP port for NLP service
	}

	// Setup gRPC server
	grpcPort := viper.GetInt("grpc.port")
	if grpcPort == 0 {
		grpcPort = 50052 // Default gRPC port for NLP service
	}

	// Create HTTP server
	httpServer := setupHTTPServer(httpHandler, nlpService, httpPort)

	// Create gRPC server
	grpcServer := setupGRPCServer(grpcHandler, nlpGRPCHandler, grpcPort)

	// Start servers concurrently
	var wg sync.WaitGroup

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		logrus.Infof("Starting HTTP server on port %d", httpPort)
		logrus.Infof("Health check available at: http://localhost:%d/health", httpPort)
		logrus.Infof("NLP API available at: http://localhost:%d/api/v1/nlp", httpPort)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			logrus.Fatalf("Failed to listen on gRPC port %d: %v", grpcPort, err)
		}

		logrus.Infof("Starting gRPC server on port %d", grpcPort)
		logrus.Infof("gRPC NLP Processing service available on port %d", grpcPort)

		if err := grpcServer.Serve(lis); err != nil {
			logrus.Fatalf("gRPC server failed to start: %v", err)
		}
	}()

	logrus.Info("All services started successfully! Press Ctrl+C to shutdown...")

	// Start background tasks
	go func() {
		// Refresh materialized views every hour
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := db.RefreshMaterializedViews(); err != nil {
					logrus.WithError(err).Error("Failed to refresh materialized views")
				} else {
					logrus.Debug("Materialized views refreshed successfully")
				}
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down servers...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Errorf("HTTP server forced to shutdown: %v", err)
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	logrus.Info("All servers exited")
}

func setupHTTPServer(httpHandler *handler.HTTPHandler, nlpService service.NLPProcessingService, port int) *http.Server {
	// Setup Gin router
	if viper.GetString("server.environment") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		modelStatus := nlpService.GetModelStatus()

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "nlp-processing",
			"timestamp": time.Now().UTC(),
			"database":  "connected",
			"models": gin.H{
				"finbert_loaded":     modelStatus.FinbertLoaded,
				"ner_model_loaded":   modelStatus.NerModelLoaded,
				"topic_model_loaded": modelStatus.TopicModelLoaded,
			},
			"servers": gin.H{
				"http": "running",
				"grpc": "running",
			},
		})
	})

	// Model status endpoint
	router.GET("/models/status", func(c *gin.Context) {
		modelStatus := nlpService.GetModelStatus()
		c.JSON(http.StatusOK, modelStatus)
	})

	// Basic database status endpoint
	router.GET("/db/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "connected",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// NLP Processing routes
		nlp := v1.Group("/nlp")
		{
			nlp.POST("/analyze", func(c *gin.Context) {
				// TODO: Implement HTTP endpoint for article analysis
				c.JSON(http.StatusNotImplemented, gin.H{
					"message": "Use gRPC endpoint for NLP processing",
				})
			})

			nlp.GET("/sentiment/trends", func(c *gin.Context) {
				// TODO: Implement HTTP endpoint for sentiment trends
				c.JSON(http.StatusNotImplemented, gin.H{
					"message": "Use gRPC endpoint for sentiment trends",
				})
			})
		}

		// Article routes (for compatibility)
		articles := v1.Group("/articles")
		{
			articles.POST("", httpHandler.CreateArticle)
			articles.GET("/:id", httpHandler.GetArticle)
			articles.GET("", httpHandler.ListArticles)
			articles.PUT("/:id/status", httpHandler.UpdateArticleStatus)
		}

		// Source routes (for compatibility)
		sources := v1.Group("/sources")
		{
			sources.POST("", httpHandler.CreateSource)
			sources.GET("/:id", httpHandler.GetSource)
			sources.GET("", httpHandler.ListSources)
			sources.PUT("/:id", httpHandler.UpdateSource)
		}

		// Ingestion routes (for compatibility)
		ingestion := v1.Group("/ingestion")
		{
			ingestion.POST("/trigger", httpHandler.TriggerManualIngestion)
			ingestion.GET("/status", httpHandler.GetIngestionStatus)
		}
	}

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
}

func setupGRPCServer(grpcHandler *handler.GRPCHandler, nlpHandler *handler.NLPGRPCHandler, port int) *grpc.Server {
	// Create gRPC server with options
	maxMsgSize := viper.GetInt("grpc.max_receive_message_size")
	if maxMsgSize == 0 {
		maxMsgSize = 1024 * 1024 * 4 // 4MB default
	}

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	}

	grpcServer := grpc.NewServer(opts...)

	// Register services with the generated proto registration functions
	newsv1.RegisterNewsServiceServer(grpcServer, grpcHandler)
	nlpv1.RegisterNLPProcessingServiceServer(grpcServer, nlpHandler)

	// Enable reflection for development
	if viper.GetString("server.environment") != "production" {
		reflection.Register(grpcServer)
		logrus.Info("gRPC reflection enabled for development")
	}

	return grpcServer
}
func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./")
	viper.AddConfigPath(".")

	// Set default values
	viper.SetDefault("server.port", 4002)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("grpc.port", 50052)
	viper.SetDefault("grpc.max_receive_message_size", 4194304)
	viper.SetDefault("grpc.max_send_message_size", 4194304)

	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.database", "nlp_processing")
	viper.SetDefault("database.postgres.username", "postgres")
	viper.SetDefault("database.postgres.password", "zakaria")
	viper.SetDefault("database.postgres.ssl_mode", "disable")
	viper.SetDefault("database.postgres.max_open_conns", 30)
	viper.SetDefault("database.postgres.max_idle_conns", 10)
	viper.SetDefault("database.postgres.conn_max_lifetime", "5m")
	viper.SetDefault("database.postgres.conn_max_idle_time", "2m")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.database", 1)
	viper.SetDefault("redis.password", "")

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	viper.SetDefault("models.finbert.model_path", "./models/finbert/model.bin")
	viper.SetDefault("models.finbert.device", "cpu")
	viper.SetDefault("models.finbert.batch_size", 32)
	viper.SetDefault("models.finbert.max_length", 512)

	viper.SetDefault("models.ner.model_path", "./models/ner/model.bin")
	viper.SetDefault("models.ner.spacy_model", "en_core_web_sm")
	viper.SetDefault("models.ner.confidence_threshold", 0.7)

	viper.SetDefault("models.topic_classifier.model_path", "./models/topic/classifier.bin")
	viper.SetDefault("models.topic_classifier.categories", []string{
		"Market Analysis", "Earnings", "Mergers & Acquisitions", "Economic Indicators",
		"Regulatory News", "Company News", "Crypto", "Commodities",
	})

	viper.SetDefault("processing.batch_size", 50)
	viper.SetDefault("processing.worker_count", 3)
	viper.SetDefault("processing.retry_attempts", 3)
	viper.SetDefault("processing.timeout_seconds", 60)

	// Read environment variables with prefix
	viper.SetEnvPrefix("NLP")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logrus.Warnf("Config file not found, using defaults: %v", err)
	} else {
		logrus.Infof("Using config file: %s", viper.ConfigFileUsed())
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
