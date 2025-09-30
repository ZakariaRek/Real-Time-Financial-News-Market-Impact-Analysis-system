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
	"google.golang.org/grpc/reflection"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/database"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/handler"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/repository"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/service"
	nlpv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/proto/gen/nlp/v1"
)

func main() {
	// Initialize configuration
	if err := initConfig(); err != nil {
		logrus.Fatalf("Failed to initialize config: %v", err)
	}

	initLogger()
	logrus.Info("Starting S&P 500 NLP Processing Service (Sentiment Analysis)...")

	// Initialize database
	db, err := initDatabase()
	if err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		logrus.Fatalf("Failed to run migrations: %v", err)
	}

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
	if err := nlpService.InitializeModels(ctx); err != nil {
		logrus.Fatalf("Failed to initialize NLP models: %v", err)
	}

	// Initialize gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(viper.GetInt("grpc.max_receive_message_size")),
		grpc.MaxSendMsgSize(viper.GetInt("grpc.max_send_message_size")),
	)

	// Register gRPC handler
	nlpHandler := handler.NewNLPGRPCHandler(nlpService, analysisRepo)
	nlpv1.RegisterNLPProcessingServiceServer(grpcServer, nlpHandler)

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

	logrus.Info("S&P 500 NLP Processing Service started successfully!")
	logrus.Info("Service is ready to analyze sentiment for S&P 500 news articles")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down NLP Processing Service...")

	// Graceful shutdown
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

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("server.port", 4002)
	viper.SetDefault("grpc.port", 50052)
	viper.SetDefault("grpc.max_receive_message_size", 4194304) // 4MB
	viper.SetDefault("grpc.max_send_message_size", 4194304)
	viper.SetDefault("processing.worker_count", 3)
	viper.SetDefault("processing.batch_size", 50)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.database", 1)

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
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.database"),
	})
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
