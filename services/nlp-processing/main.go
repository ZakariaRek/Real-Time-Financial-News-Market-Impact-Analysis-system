// services/nlp-processing/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
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

// ServiceContainer holds all initialized services
type ServiceContainer struct {
	analysisRepo     repository.AnalysisRepository
	cacheRepo        repository.CacheRepository
	sentimentService service.SentimentService
	nlpService       service.NLPProcessingService
	nlpHandler       *handler.NLPGRPCHandler
	streamProcessor  *service.StreamProcessor
	db               *database.Database
	redisClient      *redis.Client
	grpcServer       *grpc.Server
	healthServer     *health.Server
}

// Metrics tracks service performance
type Metrics struct {
	ArticlesProcessed int64
	ArticlesSucceeded int64
	ArticlesFailed    int64
	TotalProcessingMs int64
	LastProcessing    time.Time
	mu                sync.RWMutex
}

var (
	serviceMetrics = &Metrics{}
	startTime      = time.Now()
	isReady        int32
	dbConnected    int32
	grpcStarted    int32
	modelsLoaded   int32
)

type HealthResponse struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Checks  map[string]string `json:"checks,omitempty"`
}

func main() {
	// Initialize configuration first
	if err := initConfig(); err != nil {
		logrus.Fatalf("Failed to initialize config: %v", err)
	}

	// Initialize logger based on config
	initLogger()

	logrus.Info("========================================")
	logrus.Info("   NLP Processing Service Starting")
	logrus.Info("========================================")
	logrus.Infof("Environment: %s", viper.GetString("server.environment"))
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
	httpPort := viper.GetInt("server.http_port")
	if httpPort == 0 {
		httpPort = 8080
	}
	grpcPort := viper.GetInt("grpc.port")

	// Start HTTP health server
	httpServer := setupHTTPServer(httpPort, container.db)

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
		startGRPCServer(container.grpcServer, grpcPort)
	}()

	// Start stream processor
	if container.streamProcessor != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logrus.Info("🚀 Starting stream processor...")
			if err := container.streamProcessor.Start(ctx); err != nil {
				logrus.WithError(err).Error("Stream processor failed")
			}
		}()
	}

	// Report metrics periodically
	wg.Add(1)
	go func() {
		defer wg.Done()
		reportMetricsPeriodically(ctx)
	}()

	// Mark service as ready
	atomic.StoreInt32(&isReady, 1)

	logrus.Info("========================================")
	logrus.Info("   All Services Running Successfully")
	logrus.Info("========================================")
	logrus.Infof("   HTTP Server:  http://localhost:%d", httpPort)
	logrus.Infof("   gRPC Server:  localhost:%d", grpcPort)
	logrus.Infof("   Health Check: http://localhost:%d/health", httpPort)
	logrus.Infof("   Ready Check:  http://localhost:%d/ready", httpPort)
	logrus.Infof("   Metrics:      http://localhost:%d/metrics", httpPort)
	logrus.Info("========================================")
	logrus.Info("   Press Ctrl+C to shutdown")
	logrus.Info("========================================")

	gracefulShutdown(httpServer, container, cancel)
	wg.Wait()
	logrus.Info("All services stopped gracefully")
}

func initConfig() error {
	// Set config file name and type
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add config paths
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/app/config") // For Docker

	// Enable environment variable reading
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
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

		// Expand environment variables in config
		if err := expandEnvironmentVariables(); err != nil {
			return fmt.Errorf("failed to expand environment variables: %w", err)
		}
	}

	// Validate critical configuration
	if err := validateConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Log configuration summary
	logConfigSummary()

	return nil
}

func setConfigDefaults() {
	viper.SetDefault("server.port", 4002)
	viper.SetDefault("server.http_port", 8080)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("grpc.port", 50052)
	viper.SetDefault("grpc.max_receive_message_size", 4194304)
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
	viper.SetDefault("database.postgres.conn_max_lifetime", "5m")
	viper.SetDefault("database.postgres.conn_max_idle_time", "2m")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("external_services.news_ingestion_service.grpc_endpoint", "localhost:50051")
	viper.SetDefault("external_services.news_ingestion_service.timeout", "30s")
}

// expandEnvironmentVariables expands ${VAR:default} syntax in all config values
func expandEnvironmentVariables() error {
	// Regular expression to match ${VAR:default} or ${VAR}
	envVarPattern := regexp.MustCompile(`\$\{([^}:]+)(?::([^}]*))?\}`)

	// Get all settings
	settings := viper.AllSettings()

	// Process each setting recursively
	expanded := expandMap(settings, envVarPattern)

	// Merge expanded values back into viper
	for key, value := range flattenMap("", expanded) {
		viper.Set(key, value)
	}

	return nil
}

// expandMap recursively expands environment variables in a map
func expandMap(m map[string]interface{}, pattern *regexp.Regexp) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range m {
		switch v := value.(type) {
		case string:
			result[key] = expandString(v, pattern)
		case map[string]interface{}:
			result[key] = expandMap(v, pattern)
		case []interface{}:
			result[key] = expandSlice(v, pattern)
		default:
			result[key] = value
		}
	}

	return result
}

// expandSlice recursively expands environment variables in a slice
func expandSlice(s []interface{}, pattern *regexp.Regexp) []interface{} {
	result := make([]interface{}, len(s))

	for i, value := range s {
		switch v := value.(type) {
		case string:
			result[i] = expandString(v, pattern)
		case map[string]interface{}:
			result[i] = expandMap(v, pattern)
		case []interface{}:
			result[i] = expandSlice(v, pattern)
		default:
			result[i] = value
		}
	}

	return result
}

// expandString replaces ${VAR:default} with environment variable value or default
func expandString(s string, pattern *regexp.Regexp) string {
	return pattern.ReplaceAllStringFunc(s, func(match string) string {
		submatches := pattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		varName := submatches[1]
		defaultValue := ""
		if len(submatches) > 2 {
			defaultValue = submatches[2]
		}

		if envValue := os.Getenv(varName); envValue != "" {
			return envValue
		}

		return defaultValue
	})
}

// flattenMap converts nested map to flat map with dot notation keys
func flattenMap(prefix string, m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range m {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			for k, val := range flattenMap(fullKey, v) {
				result[k] = val
			}
		default:
			result[fullKey] = value
		}
	}

	return result
}

func validateConfig() error {
	if viper.GetString("database.postgres.host") == "" {
		return fmt.Errorf("database host is required")
	}

	if viper.GetInt("database.postgres.port") == 0 {
		return fmt.Errorf("database port is required")
	}

	if viper.GetString("database.postgres.database") == "" {
		return fmt.Errorf("database name is required")
	}

	if viper.GetString("database.postgres.username") == "" {
		return fmt.Errorf("database username is required")
	}

	if viper.GetInt("grpc.port") == 0 {
		return fmt.Errorf("gRPC port is required")
	}

	return nil
}

func logConfigSummary() {
	logrus.Info("Configuration Summary:")
	logrus.Infof("  HTTP Port: %d", viper.GetInt("server.http_port"))
	logrus.Infof("  gRPC Port: %d", viper.GetInt("grpc.port"))
	logrus.Infof("  Environment: %s", viper.GetString("server.environment"))
	logrus.Infof("  Database Host: %s", viper.GetString("database.postgres.host"))
	logrus.Infof("  Database Port: %d", viper.GetInt("database.postgres.port"))
	logrus.Infof("  Database Name: %s", viper.GetString("database.postgres.database"))
	logrus.Infof("  Redis Host: %s", viper.GetString("redis.host"))
	logrus.Infof("  Redis Port: %d", viper.GetInt("redis.port"))
	logrus.Infof("  Log Level: %s", viper.GetString("logging.level"))
	logrus.Infof("  Worker Count: %d", viper.GetInt("processing.worker_count"))
	logrus.Infof("  Batch Size: %d", viper.GetInt("processing.batch_size"))

	nlpEndpoint := viper.GetString("external_services.news_ingestion_service.grpc_endpoint")
	logrus.Infof("  News Ingestion: %s", nlpEndpoint)
}

func initLogger() {
	level := viper.GetString("logging.level")
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

	format := viper.GetString("logging.format")
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

	logrus.Infof("Connecting to database at %s:%d/%s",
		dbConfig.Host, dbConfig.Port, dbConfig.Database)

	var db *database.Database
	var err error

	// Retry connection with exponential backoff
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = database.NewConnection(dbConfig)
		if err == nil {
			logrus.Info("✓ Database connection established")
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

func initRedis() *redis.Client {
	password := viper.GetString("redis.password")

	opts := &redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			viper.GetString("redis.host"),
			viper.GetInt("redis.port")),
		DB: viper.GetInt("redis.database"),
	}

	if password != "" {
		opts.Password = password
		logrus.Info("Connecting to Redis with authentication")
	} else {
		logrus.Info("Connecting to Redis without authentication")
	}

	return redis.NewClient(opts)
}

func initializeServices() (*ServiceContainer, error) {
	logrus.Info("Initializing services...")

	// Initialize database
	db, err := initDatabase()
	if err != nil {
		return nil, fmt.Errorf("database initialization failed: %w", err)
	}
	atomic.StoreInt32(&dbConnected, 1)

	// Run migrations
	logrus.Info("Starting database migrations...")
	if err := db.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	logrus.Info("✓ Database migrations completed")

	// Create indexes
	if err := db.CreateIndexes(); err != nil {
		logrus.Warnf("Some indexes failed to create: %v", err)
	} else {
		logrus.Info("✓ Database indexes created")
	}

	// Initialize Redis
	redisClient := initRedis()
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logrus.Warnf("Redis connection failed, continuing without cache: %v", err)
	} else {
		logrus.Info("✓ Redis connected successfully")
	}

	// Initialize repositories
	analysisRepo := repository.NewAnalysisRepository(db.DB)
	cacheRepo := repository.NewCacheRepository(redisClient)
	logrus.Info("✓ Repositories initialized")

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
		return nil, fmt.Errorf("failed to initialize NLP models: %w", err)
	}
	logrus.Info("✓ NLP models loaded successfully")
	atomic.StoreInt32(&modelsLoaded, 1)

	// Initialize handlers
	nlpHandler := handler.NewNLPGRPCHandler(nlpService, analysisRepo)
	logrus.Info("✓ gRPC handler initialized")

	// Setup gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(viper.GetInt("grpc.max_receive_message_size")),
		grpc.MaxSendMsgSize(viper.GetInt("grpc.max_send_message_size")),
	)

	// Register services
	nlpv1.RegisterNLPProcessingServiceServer(grpcServer, nlpHandler)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("nlp.v1.NLPProcessingService", grpc_health_v1.HealthCheckResponse_SERVING)
	logrus.Info("✓ gRPC Health Check service registered")

	// Enable reflection
	if viper.GetString("server.environment") != "production" {
		reflection.Register(grpcServer)
		logrus.Info("✓ gRPC reflection enabled")
	}

	// Initialize stream processor
	var streamProcessor *service.StreamProcessor
	newsIngestionEndpoint := viper.GetString("external_services.news_ingestion_service.grpc_endpoint")
	if newsIngestionEndpoint != "" {
		config := service.StreamProcessorConfig{
			NewsIngestionEndpoint: newsIngestionEndpoint,
			WorkerCount:           viper.GetInt("processing.worker_count"),
			BatchSize:             int32(viper.GetInt("processing.batch_size")),
		}

		processor, err := service.NewStreamProcessor(config, nlpService)
		if err != nil {
			logrus.Warnf("⚠️ Failed to initialize stream processor: %v", err)
			logrus.Info("Service will run without automatic article processing")
		} else {
			streamProcessor = processor
			logrus.Info("✓ Stream processor initialized")
		}
	}

	return &ServiceContainer{
		analysisRepo:     analysisRepo,
		cacheRepo:        cacheRepo,
		sentimentService: sentimentService,
		nlpService:       nlpService,
		nlpHandler:       nlpHandler,
		streamProcessor:  streamProcessor,
		db:               db,
		redisClient:      redisClient,
		grpcServer:       grpcServer,
		healthServer:     healthServer,
	}, nil
}

func (c *ServiceContainer) cleanup() {
	logrus.Info("🧹 Cleaning up resources...")

	if c.redisClient != nil {
		if err := c.redisClient.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close Redis connection")
		} else {
			logrus.Info("✓ Closed Redis connection")
		}
	}

	if c.streamProcessor != nil {
		if err := c.streamProcessor.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close stream processor")
		} else {
			logrus.Info("✓ Closed stream processor")
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

func setupHTTPServer(port int, db *database.Database) *http.Server {
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":    "healthy",
			"service":   "nlp-processing",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC(),
			"database":  getDatabaseStatus(db),
			"servers": map[string]string{
				"http": "running",
				"grpc": "running",
			},
			"metrics": getMetrics(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// Ready endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		checks := make(map[string]string)
		allReady := true

		if atomic.LoadInt32(&dbConnected) == 1 {
			checks["database"] = "connected"
		} else {
			checks["database"] = "disconnected"
			allReady = false
		}

		if atomic.LoadInt32(&modelsLoaded) == 1 {
			checks["models"] = "loaded"
		} else {
			checks["models"] = "loading"
			allReady = false
		}

		if atomic.LoadInt32(&grpcStarted) == 1 {
			checks["grpc"] = "started"
		} else {
			checks["grpc"] = "starting"
			allReady = false
		}

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

	// Live endpoint
	mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:  "alive",
			Service: "nlp-processing",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// Metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(getMetrics())
	})

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func startHTTPServer(server *http.Server, port int) {
	logrus.Infof("🌐 Starting HTTP server on port %d", port)
	logrus.Infof("  → Health:  http://localhost:%d/health", port)
	logrus.Infof("  → Ready:   http://localhost:%d/ready", port)
	logrus.Infof("  → Live:    http://localhost:%d/live", port)
	logrus.Infof("  → Metrics: http://localhost:%d/metrics", port)

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

	atomic.StoreInt32(&grpcStarted, 1)

	if err := server.Serve(lis); err != nil {
		logrus.Fatalf("gRPC server failed: %v", err)
	}
}

func gracefulShutdown(httpServer *http.Server, container *ServiceContainer, cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("========================================")
	logrus.Info("   Shutdown signal received")
	logrus.Info("========================================")

	atomic.StoreInt32(&isReady, 0)

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
	stopped := make(chan struct{})
	go func() {
		container.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		logrus.Info("✓ gRPC server stopped gracefully")
	case <-time.After(30 * time.Second):
		logrus.Warn("Forcing gRPC server shutdown")
		container.grpcServer.Stop()
	}

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
				"articles_processed": metrics["articles_processed"],
				"articles_succeeded": metrics["articles_succeeded"],
				"articles_failed":    metrics["articles_failed"],
				"avg_processing_ms":  metrics["avg_processing_ms"],
			}).Info("📊 Service metrics report")
		}
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

	avgProcessingMs := int64(0)
	if serviceMetrics.ArticlesProcessed > 0 {
		avgProcessingMs = serviceMetrics.TotalProcessingMs / serviceMetrics.ArticlesProcessed
	}

	successRate := float64(0)
	if serviceMetrics.ArticlesProcessed > 0 {
		successRate = float64(serviceMetrics.ArticlesSucceeded) / float64(serviceMetrics.ArticlesProcessed) * 100
	}

	return map[string]interface{}{
		"articles_processed": serviceMetrics.ArticlesProcessed,
		"articles_succeeded": serviceMetrics.ArticlesSucceeded,
		"articles_failed":    serviceMetrics.ArticlesFailed,
		"success_rate":       fmt.Sprintf("%.2f%%", successRate),
		"avg_processing_ms":  avgProcessingMs,
		"last_processing":    serviceMetrics.LastProcessing.Format(time.RFC3339),
		"uptime":             time.Since(startTime).String(),
	}
}

func updateProcessingMetrics(succeeded bool, processingMs int64) {
	serviceMetrics.mu.Lock()
	defer serviceMetrics.mu.Unlock()

	serviceMetrics.ArticlesProcessed++
	serviceMetrics.TotalProcessingMs += processingMs
	serviceMetrics.LastProcessing = time.Now()

	if succeeded {
		serviceMetrics.ArticlesSucceeded++
	} else {
		serviceMetrics.ArticlesFailed++
	}
}
