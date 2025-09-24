// services/news-ingestion/main.go
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

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/client"
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

	logrus.Info("Starting News Ingestion Service...")

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

	// Initialize repositories
	articleRepo := repository.NewArticleRepository(db.DB)
	sourceRepo := repository.NewSourceRepository(db.DB)
	logRepo := repository.NewProcessingLogRepository(db.DB)
	rateLimitRepo := repository.NewRateLimitRepository(db.DB)

	// Initialize clients
	newsAPIClient := client.NewNewsAPIClient(
		viper.GetString("news_sources.newsapi.api_key"),
		viper.GetString("news_sources.newsapi.base_url"),
	)

	rssClient := client.NewRSSClient()

	twitterClient := client.NewTwitterClient(
		viper.GetString("news_sources.twitter.bearer_token"),
		viper.GetString("news_sources.twitter.base_url"),
	)

	// Initialize services
	deduplicationService := service.NewDeduplicationService(articleRepo)

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

	// Initialize handlers
	httpHandler := handler.NewHTTPHandler(ingestionService, articleRepo, sourceRepo)
	grpcHandler := handler.NewGRPCHandler(ingestionService, articleRepo, sourceRepo, logRepo)

	// Setup HTTP server
	httpPort := viper.GetInt("server.port")
	if httpPort == 0 {
		httpPort = 4001 // Default HTTP port
	}

	// Setup gRPC server
	grpcPort := viper.GetInt("server.grpc_port")
	if grpcPort == 0 {
		grpcPort = 4002 // Default gRPC port
	}

	// Create HTTP server
	httpServer := setupHTTPServer(httpHandler, httpPort)

	// Create gRPC server
	grpcServer := setupGRPCServer(grpcHandler, grpcPort)

	// Start servers concurrently
	var wg sync.WaitGroup

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		logrus.Infof("Starting HTTP server on port %d", httpPort)
		logrus.Infof("Health check available at: http://localhost:%d/health", httpPort)
		logrus.Infof("Database status available at: http://localhost:%d/db/status", httpPort)

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

		if err := grpcServer.Serve(lis); err != nil {
			logrus.Fatalf("gRPC server failed to start: %v", err)
		}
	}()

	logrus.Info("All services started successfully! Press Ctrl+C to shutdown...")

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

func setupHTTPServer(httpHandler *handler.HTTPHandler, port int) *http.Server {
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
			"database":  "connected",
			"servers": gin.H{
				"http": "running",
				"grpc": "running",
			},
		})
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
		}
	}

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
}

func setupGRPCServer(grpcHandler *handler.GRPCHandler, port int) *grpc.Server {
	// Create gRPC server with options
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 4), // 4MB
		grpc.MaxSendMsgSize(1024 * 1024 * 4), // 4MB
	}

	grpcServer := grpc.NewServer(opts...)

	// Register the service
	// Note: You'll need to import the generated protobuf code
	// newsv1.RegisterNewsServiceServer(grpcServer, grpcHandler)

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
	viper.AddConfigPath(".")
	viper.AddConfigPath("./")

	// Set default values
	viper.SetDefault("server.port", 4001)
	viper.SetDefault("server.grpc_port", 4002)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.database", "news_ingestion")
	viper.SetDefault("database.postgres.username", "postgres")
	viper.SetDefault("database.postgres.password", "yahyasd56") // zakaria
	viper.SetDefault("database.postgres.ssl_mode", "disable")
	viper.SetDefault("database.postgres.max_open_conns", 25)
	viper.SetDefault("database.postgres.max_idle_conns", 5)
	viper.SetDefault("database.postgres.conn_max_lifetime", "5m")
	viper.SetDefault("database.postgres.conn_max_idle_time", "1m")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "text")
	viper.SetDefault("news_sources.newsapi.api_key", "")
	viper.SetDefault("news_sources.newsapi.base_url", "https://newsapi.org/v2")
	viper.SetDefault("news_sources.twitter.bearer_token", "")
	viper.SetDefault("news_sources.twitter.base_url", "https://api.twitter.com/2")

	// Read environment variables with prefix
	viper.SetEnvPrefix("NEWS")
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
