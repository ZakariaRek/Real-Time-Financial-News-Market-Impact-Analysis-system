// services/nlp-processing/cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/database"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/handler"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/repository"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/service"
	nlpv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/proto/gen/nlp/v1"
)

func main() {
	// Load configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("Error reading config file: %v", err)
	}

	// Setup logging
	setupLogging()

	logrus.Info("🚀 Starting NLP Processing Service...")

	// Initialize database
	db, err := initializeDatabase()
	if err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient := initializeRedis()
	defer redisClient.Close()

	// Initialize repositories
	analysisRepo := repository.NewAnalysisRepository(db.DB)
	cacheRepo := repository.NewCacheRepository(redisClient)
	articleRepo := repository.NewArticleRepository(db.DB)
	sourceRepo := repository.NewSourceRepository(db.DB)
	processingLogRepo := repository.NewProcessingLogRepository(db.DB)

	// Initialize ML services
	sentimentService := initializeSentimentService()
	nerService := initializeNERService()
	topicService := initializeTopicService()

	// Initialize NLP processing service
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
	ctx := context.Background()
	if err := nlpService.InitializeModels(ctx); err != nil {
		logrus.Fatalf("Failed to initialize NLP models: %v", err)
	}

	// Initialize stream processor
	streamConfig := service.StreamProcessorConfig{
		NewsIngestionEndpoint: viper.GetString("external_services.news_ingestion_service.grpc_endpoint"),
		WorkerCount:           nlpConfig.WorkerCount,
		BatchSize:             int32(nlpConfig.BatchSize),
	}

	streamProcessor, err := service.NewStreamProcessor(
		streamConfig,
		nlpService,
		articleRepo,
		processingLogRepo,
	)
	if err != nil {
		logrus.Fatalf("Failed to create stream processor: %v", err)
	}
	defer streamProcessor.Close()

	// Start gRPC server
	grpcServer := startGRPCServer(nlpService, analysisRepo, articleRepo, sourceRepo, processingLogRepo)
	defer grpcServer.GracefulStop()

	// Start stream processing in background
	processorCtx, cancelProcessor := context.WithCancel(context.Background())
	defer cancelProcessor()

	go func() {
		if err := streamProcessor.Start(processorCtx); err != nil {
			logrus.WithError(err).Error("Stream processor error")
		}
	}()

	logrus.Info("✅ NLP Processing Service is running")
	logrus.Infof("📊 Processing articles from: %s", streamConfig.NewsIngestionEndpoint)
	logrus.Infof("🔧 Workers: %d, Batch Size: %d", streamConfig.WorkerCount, streamConfig.BatchSize)

	// Wait for shutdown signal
	waitForShutdown(cancelProcessor)

	logrus.Info("🛑 NLP Processing Service stopped")
}

func setupLogging() {
	level := viper.GetString("logging.level")
	format := viper.GetString("logging.format")

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)

	if format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
}

func initializeDatabase() (*database.Database, error) {
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

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create indexes
	if err := db.CreateIndexes(); err != nil {
		logrus.WithError(err).Warn("Failed to create some indexes")
	}

	// Create materialized views
	if err := db.CreateMaterializedViews(); err != nil {
		logrus.WithError(err).Warn("Failed to create materialized views")
	}

	return db, nil
}

func initializeRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.database"),
		PoolSize:     viper.GetInt("redis.pool_size"),
		MinIdleConns: viper.GetInt("redis.min_idle_conns"),
	})
}

func initializeSentimentService() service.SentimentService {
	config := service.SentimentConfig{
		ModelPath: viper.GetString("models.finbert.model_path"),
		Device:    viper.GetString("models.finbert.device"),
		BatchSize: viper.GetInt("models.finbert.batch_size"),
		MaxLength: viper.GetInt("models.finbert.max_length"),
	}
	return service.NewSentimentService(config)
}

func initializeNERService() service.NERService {
	config := service.NERConfig{
		ModelPath:           viper.GetString("models.ner.model_path"),
		SpacyModel:          viper.GetString("models.ner.spacy_model"),
		ConfidenceThreshold: float32(viper.GetFloat64("models.ner.confidence_threshold")),
	}
	return service.NewNERService(config)
}

func initializeTopicService() service.TopicService {
	config := service.TopicConfig{
		ModelPath:  viper.GetString("models.topic_classifier.model_path"),
		Categories: viper.GetStringSlice("models.topic_classifier.categories"),
	}
	return service.NewTopicService(config)
}

func startGRPCServer(
	nlpService service.NLPProcessingService,
	analysisRepo repository.AnalysisRepository,
	articleRepo repository.ArticleRepository,
	sourceRepo repository.SourceRepository,
	processingLogRepo repository.ProcessingLogRepository,
) *grpc.Server {
	port := viper.GetInt("grpc.port")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logrus.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(viper.GetInt("grpc.max_receive_message_size")),
		grpc.MaxSendMsgSize(viper.GetInt("grpc.max_send_message_size")),
	)

	// Register NLP handler
	nlpHandler := handler.NewNLPGRPCHandler(nlpService, analysisRepo)
	nlpv1.RegisterNLPProcessingServiceServer(grpcServer, nlpHandler)

	go func() {
		logrus.Infof("🌐 gRPC server listening on port %d", port)
		if err := grpcServer.Serve(lis); err != nil {
			logrus.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	return grpcServer
}

func waitForShutdown(cancelProcessor context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logrus.Info("Shutdown signal received, stopping gracefully...")

	// Cancel stream processor
	cancelProcessor()

	// Give it some time to finish
	time.Sleep(5 * time.Second)
}
