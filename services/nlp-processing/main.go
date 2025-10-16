// services/nlp-processing/cmd/server/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/database"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/handler"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/repository"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/service"
	nlpv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/proto/gen/nlp/v1"
)

// Health status tracking
var (
	isReady      int32 // 0 = not ready, 1 = ready
	dbConnected  int32
	grpcStarted  int32
	modelsLoaded int32
)

type HealthResponse struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Checks  map[string]string `json:"checks,omitempty"`
}

func main() {
	// Initialize configuration
	if err := initConfig(); err != nil {
		logrus.Fatalf("Failed to initialize config: %v", err)
	}

	initLogger()
	logrus.Info("Starting S&P 500 NLP Processing Service (Sentiment Analysis)...")

	// Start HTTP health server IMMEDIATELY (before migrations)
	httpPort := viper.GetInt("server.http_port")
	if httpPort == 0 {
		httpPort = 8080 // default
	}

	httpServer := startHealthServer(httpPort)
	logrus.Infof("HTTP health server started on port %d", httpPort)

	// Initialize database
	db, err := initDatabase()
	if err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	atomic.StoreInt32(&dbConnected, 1)

	// Run migrations
	logrus.Info("Starting database migrations...")
	if err := db.AutoMigrate(); err != nil {
		logrus.Fatalf("Failed to run migrations: %v", err)
	}
	logrus.Info("Migrations completed successfully")

	// Create indexes for better query performance
	if err := db.CreateIndexes(); err != nil {
		logrus.Warnf("Failed to create some indexes: %v", err)
	}

	// Initialize Redis
	redisClient := initRedis()
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logrus.Warnf("Redis connection failed, continuing without cache: %v", err)
	} else {
		logrus.Info("Redis connected successfully")
	}

	// Initialize repositories
	analysisRepo := repository.NewAnalysisRepository(db.DB)
	cacheRepo := repository.NewCacheRepository(redisClient)

	// Initialize services
	logrus.Info("Initializing sentiment analysis service...")
	sentimentService := service.NewSentimentService()

	nlpService := service.NewNLPProcessingService(
		sentimentService,
		analysisRepo,
		cacheRepo,
	)

	// Initialize models
	ctx := context.Background()
	logrus.Info("Loading NLP models...")
	if err := nlpService.InitializeModels(ctx); err != nil {
		logrus.Fatalf("Failed to initialize NLP models: %v", err)
	}
	logrus.Info("NLP models loaded successfully")
	atomic.StoreInt32(&modelsLoaded, 1)

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(viper.GetInt("grpc.max_receive_message_size")),
		grpc.MaxSendMsgSize(viper.GetInt("grpc.max_send_message_size")),
	)

	// Register gRPC handler
	nlpHandler := handler.NewNLPGRPCHandler(nlpService, analysisRepo)
	nlpv1.RegisterNLPProcessingServiceServer(grpcServer, nlpHandler)

	// Register standard gRPC Health Check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("nlp.v1.NLPProcessingService", grpc_health_v1.HealthCheckResponse_SERVING)
	logrus.Info("gRPC Health Check service registered")

	// Enable reflection for grpcurl
	reflection.Register(grpcServer)

	// Start gRPC server
	grpcPort := viper.GetInt("grpc.port")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		logrus.Fatalf("Failed to listen on port %d: %v", grpcPort, err)
	}

	go func() {
		logrus.Infof("gRPC server listening on port %d", grpcPort)
		atomic.StoreInt32(&grpcStarted, 1)
		if err := grpcServer.Serve(lis); err != nil {
			logrus.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Initialize and start stream processor
	streamProcessor, err := initStreamProcessor(nlpService)
	if err != nil {
		logrus.Warnf("Failed to initialize stream processor: %v", err)
		logrus.Info("Service will run without automatic article processing")
	} else {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			if err := streamProcessor.Start(ctx); err != nil {
				logrus.WithError(err).Error("Stream processor failed")
			}
		}()
	}

	// Mark service as ready
	atomic.StoreInt32(&isReady, 1)
	logrus.Info("S&P 500 NLP Processing Service started successfully!")
	logrus.Info("Service is ready to analyze sentiment for S&P 500 news articles")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down NLP Processing Service...")

	// Mark as not ready during shutdown
	atomic.StoreInt32(&isReady, 0)

	// Graceful shutdown of HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Warnf("HTTP server shutdown error: %v", err)
	}

	// Graceful shutdown of gRPC server
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		logrus.Info("gRPC server stopped gracefully")
	case <-time.After(30 * time.Second):
		logrus.Warn("Forcing gRPC server shutdown")
		grpcServer.Stop()
	}

	if streamProcessor != nil {
		streamProcessor.Close()
	}

	logrus.Info("NLP Processing Service shutdown complete")
}

// startHealthServer starts an HTTP server for health checks
func startHealthServer(port int) *http.Server {
	mux := http.NewServeMux()

	// Basic health endpoint - returns 200 if server is running
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:  "healthy",
			Service: "nlp-processing",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// Readiness endpoint - checks if service is fully initialized
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		checks := make(map[string]string)
		allReady := true

		// Check database
		if atomic.LoadInt32(&dbConnected) == 1 {
			checks["database"] = "connected"
		} else {
			checks["database"] = "disconnected"
			allReady = false
		}

		// Check models
		if atomic.LoadInt32(&modelsLoaded) == 1 {
			checks["models"] = "loaded"
		} else {
			checks["models"] = "loading"
			allReady = false
		}

		// Check gRPC server
		if atomic.LoadInt32(&grpcStarted) == 1 {
			checks["grpc"] = "started"
		} else {
			checks["grpc"] = "starting"
			allReady = false
		}

		// Overall readiness
		if atomic.LoadInt32(&isReady) == 1 {
			checks["overall"] = "ready"
		} else {
			checks["overall"] = "not_ready"
			allReady = false
		}

		response := HealthResponse{
			Service: "nlp-processing",
			Checks:  checks,
		}

		w.Header().Set("Content-Type", "application/json")

		if allReady {
			response.Status = "ready"
			w.WriteHeader(http.StatusOK)
		} else {
			response.Status = "not_ready"
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(response)
	})

	// Liveness endpoint - simple check that process is alive
	mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:  "alive",
			Service: "nlp-processing",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	return server
}

func initConfig() error {
	// Set up viper to read from config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Enable reading from environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	viper.SetDefault("server.port", 4002)
	viper.SetDefault("server.http_port", 8080)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("grpc.port", 50052)
	viper.SetDefault("grpc.max_receive_message_size", 4194304) // 4MB
	viper.SetDefault("grpc.max_send_message_size", 4194304)
	viper.SetDefault("processing.worker_count", 3)
	viper.SetDefault("processing.batch_size", 50)
	viper.SetDefault("processing.retry_attempts", 3)
	viper.SetDefault("processing.timeout_seconds", 60)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.database", 1)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.database", "nlp_processing")
	viper.SetDefault("database.postgres.username", "postgres")
	viper.SetDefault("database.postgres.password", "postgres")
	viper.SetDefault("database.postgres.ssl_mode", "disable")
	viper.SetDefault("database.postgres.max_open_conns", 30)
	viper.SetDefault("database.postgres.max_idle_conns", 10)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("external_services.news_ingestion_service.grpc_endpoint", "localhost:50051")
	viper.SetDefault("external_services.news_ingestion_service.timeout", "30s")

	// Try to read config file (optional - env vars will override)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logrus.Warn("Config file not found, using environment variables and defaults")
		} else {
			logrus.Warnf("Error reading config file: %v, using environment variables and defaults", err)
		}
	} else {
		logrus.Infof("Using config file: %s", viper.ConfigFileUsed())
	}

	// Bind environment variables explicitly
	bindEnvVars()

	return nil
}

func bindEnvVars() {
	// Server
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.http_port", "SERVER_HTTP_PORT")
	viper.BindEnv("server.environment", "ENVIRONMENT")

	// Database
	viper.BindEnv("database.postgres.host", "POSTGRES_HOST")
	viper.BindEnv("database.postgres.port", "POSTGRES_PORT")
	viper.BindEnv("database.postgres.database", "POSTGRES_DB")
	viper.BindEnv("database.postgres.username", "POSTGRES_USER")
	viper.BindEnv("database.postgres.password", "POSTGRES_PASSWORD")
	viper.BindEnv("database.postgres.ssl_mode", "DATABASE_SSL_MODE")
	viper.BindEnv("database.postgres.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	viper.BindEnv("database.postgres.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	viper.BindEnv("database.postgres.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME")
	viper.BindEnv("database.postgres.conn_max_idle_time", "DATABASE_CONN_MAX_IDLE_TIME")

	// Redis
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.database", "REDIS_DATABASE")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")

	// Logging
	viper.BindEnv("logging.level", "LOG_LEVEL")
	viper.BindEnv("logging.format", "LOG_FORMAT")

	// Processing
	viper.BindEnv("processing.worker_count", "PROCESSING_WORKER_COUNT")
	viper.BindEnv("processing.batch_size", "PROCESSING_BATCH_SIZE")
	viper.BindEnv("processing.retry_attempts", "PROCESSING_RETRY_ATTEMPTS")
	viper.BindEnv("processing.timeout_seconds", "PROCESSING_TIMEOUT_SECONDS")

	// gRPC
	viper.BindEnv("grpc.port", "GRPC_PORT")
	viper.BindEnv("grpc.max_receive_message_size", "GRPC_MAX_RECEIVE_MESSAGE_SIZE")
	viper.BindEnv("grpc.max_send_message_size", "GRPC_MAX_SEND_MESSAGE_SIZE")

	// External Services
	viper.BindEnv("external_services.news_ingestion_service.grpc_endpoint", "NEWS_INGESTION_ENDPOINT")
	viper.BindEnv("external_services.news_ingestion_service.timeout", "NEWS_INGESTION_TIMEOUT")
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

	if viper.GetString("logging.format") == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			DisableColors: false,
		})
	}
}

func initDatabase() (*database.Database, error) {
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
		return nil, err
	}

	logrus.Info("Database connected successfully")
	return db, nil
}

func initRedis() *redis.Client {
	password := viper.GetString("redis.password")

	opts := &redis.Options{
		Addr: fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
		DB:   viper.GetInt("redis.database"),
	}

	// Only set password if it's not empty
	if password != "" {
		opts.Password = password
		logrus.Info("Connecting to Redis with authentication")
	} else {
		logrus.Info("Connecting to Redis without authentication (no password)")
	}

	return redis.NewClient(opts)
}

func initStreamProcessor(nlpService service.NLPProcessingService) (*service.StreamProcessor, error) {
	newsIngestionEndpoint := viper.GetString("external_services.news_ingestion_service.grpc_endpoint")
	if newsIngestionEndpoint == "" {
		newsIngestionEndpoint = "localhost:50051"
	}

	config := service.StreamProcessorConfig{
		NewsIngestionEndpoint: newsIngestionEndpoint,
		WorkerCount:           viper.GetInt("processing.worker_count"),
		BatchSize:             int32(viper.GetInt("processing.batch_size")),
	}

	processor, err := service.NewStreamProcessor(config, nlpService)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream processor: %w", err)
	}

	logrus.Info("Stream processor initialized successfully")
	return processor, nil
}
