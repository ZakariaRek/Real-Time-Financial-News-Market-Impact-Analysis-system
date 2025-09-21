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
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/database"
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
		})
	})

	// Basic database status endpoint
	router.GET("/db/status", func(c *gin.Context) {
		// Test database connection
		sqlDB, err := db.DB.DB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to get database instance",
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Database ping failed",
			})
			return
		}

		stats := sqlDB.Stats()
		c.JSON(http.StatusOK, gin.H{
			"status": "connected",
			"stats": gin.H{
				"open_connections":     stats.OpenConnections,
				"in_use":               stats.InUse,
				"idle":                 stats.Idle,
				"wait_count":           stats.WaitCount,
				"wait_duration":        stats.WaitDuration.String(),
				"max_idle_closed":      stats.MaxIdleClosed,
				"max_idle_time_closed": stats.MaxIdleTimeClosed,
				"max_lifetime_closed":  stats.MaxLifetimeClosed,
			},
		})
	})

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
		logrus.Infof("Health check available at: http://localhost:%d/health", port)
		logrus.Infof("Database status available at: http://localhost:%d/db/status", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	logrus.Info("Service started successfully! Press Ctrl+C to shutdown...")

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
	viper.AddConfigPath("./")

	// Set default values
	viper.SetDefault("server.port", 4001)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.database", "news_ingestion")
	viper.SetDefault("database.postgres.username", "postgres")
	viper.SetDefault("database.postgres.password", "zakaria")
	viper.SetDefault("database.postgres.ssl_mode", "disable")
	viper.SetDefault("database.postgres.max_open_conns", 25)
	viper.SetDefault("database.postgres.max_idle_conns", 5)
	viper.SetDefault("database.postgres.conn_max_lifetime", "5m")
	viper.SetDefault("database.postgres.conn_max_idle_time", "1m")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "text")

	// Read environment variables with prefix
	viper.SetEnvPrefix("NEWS")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logrus.Warnf("Config file not found, using defaults: %v", err)
	} else {
		logrus.Infof("Using config file: %s", viper.ConfigFileUsed())
	}

	// Debug: Print database configuration
	logrus.Debugf("Database config - Host: %s, Port: %d, Database: %s, Username: %s, SSL: %s",
		viper.GetString("database.postgres.host"),
		viper.GetInt("database.postgres.port"),
		viper.GetString("database.postgres.database"),
		viper.GetString("database.postgres.username"),
		viper.GetString("database.postgres.ssl_mode"))

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
