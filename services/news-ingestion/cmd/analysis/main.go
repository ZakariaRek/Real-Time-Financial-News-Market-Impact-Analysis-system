// // services/news-ingestion/cmd/analysis/main.go
package main

//
//import (
//	"context"
//	"fmt"
//	"net/http"
//	"os"
//	"os/signal"
//	"sync"
//	"syscall"
//	"time"
//
//	"github.com/gin-gonic/gin"
//	"github.com/gorilla/websocket"
//	"github.com/sirupsen/logrus"
//	"github.com/spf13/viper"
//
//	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/analysis"
//	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/client"
//	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/database"
//	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/repository"
//	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/service"
//)
//
//func main() {
//	// Initialize configuration
//	if err := initConfig(); err != nil {
//		logrus.Fatalf("Failed to initialize config: %v", err)
//	}
//
//	initLogger()
//	logrus.Info("Starting Real-Time Financial Analysis System...")
//
//	// Initialize database
//	db, err := initDatabase()
//	if err != nil {
//		logrus.Fatalf("Failed to initialize database: %v", err)
//	}
//	defer db.Close()
//
//	// Initialize repositories
//	articleRepo := repository.NewArticleRepository(db.DB)
//	sourceRepo := repository.NewSourceRepository(db.DB)
//	logRepo := repository.NewProcessingLogRepository(db.DB)
//	rateLimitRepo := repository.NewRateLimitRepository(db.DB)
//
//	// Initialize analysis repository
//	analysisRepo := repository.NewAnalysisRepository(db.DB)
//
//	// Initialize clients
//	newsAPIClient := client.NewNewsAPIClient(
//		viper.GetString("news_sources.newsapi.api_key"),
//		viper.GetString("news_sources.newsapi.base_url"),
//	)
//	rssClient := client.NewRSSClient()
//	marketDataClient := client.NewMarketDataClient(
//		viper.GetString("market_data.alpha_vantage_key"),
//	)
//
//	// Initialize services
//	sentimentService := analysis.NewSentimentService()
//	impactPredictionService := analysis.NewImpactPredictionService(analysisRepo)
//	realtimeAnalysisService := analysis.NewRealtimeAnalysisService(
//		articleRepo,
//		sourceRepo,
//		analysisRepo,
//		sentimentService,
//		impactPredictionService,
//		marketDataClient,
//		newsAPIClient,
//		rssClient,
//	)
//
//	// Initialize WebSocket hub for real-time updates
//	wsHub := analysis.NewWebSocketHub()
//
//	// Start background services
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	var wg sync.WaitGroup
//
//	// Start real-time analysis
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		if err := realtimeAnalysisService.Start(ctx, wsHub); err != nil {
//			logrus.WithError(err).Error("Real-time analysis service failed")
//		}
//	}()
//
//	// Start WebSocket hub
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		wsHub.Run(ctx)
//	}()
//
//	// Setup HTTP server for dashboard
//	router := setupRouter(realtimeAnalysisService, wsHub)
//	server := &http.Server{
//		Addr:    fmt.Sprintf(":%d", viper.GetInt("analysis.dashboard_port")),
//		Handler: router,
//	}
//
//	// Start HTTP server
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		logrus.Infof("Starting dashboard server on port %d", viper.GetInt("analysis.dashboard_port"))
//		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
//			logrus.WithError(err).Error("Dashboard server failed")
//		}
//	}()
//
//	logrus.Info("Real-time financial analysis system started successfully!")
//
//	// Wait for interrupt signal
//	quit := make(chan os.Signal, 1)
//	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
//	<-quit
//
//	logrus.Info("Shutting down...")
//	cancel()
//
//	// Graceful shutdown
//	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer shutdownCancel()
//
//	if err := server.Shutdown(shutdownCtx); err != nil {
//		logrus.WithError(err).Error("Server forced to shutdown")
//	}
//
//	wg.Wait()
//	logrus.Info("Real-time analysis system shutdown complete")
//}
//
//func setupRouter(analysisService *analysis.RealtimeAnalysisService, wsHub *analysis.WebSocketHub) *gin.Engine {
//	if viper.GetString("server.environment") == "production" {
//		gin.SetMode(gin.ReleaseMode)
//	}
//
//	router := gin.New()
//	router.Use(gin.Logger())
//	router.Use(gin.Recovery())
//
//	// CORS middleware
//	router.Use(func(c *gin.Context) {
//		c.Header("Access-Control-Allow-Origin", "*")
//		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
//		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
//		if c.Request.Method == "OPTIONS" {
//			c.AbortWithStatus(204)
//			return
//		}
//		c.Next()
//	})
//
//	// Serve static files
//	router.Static("/static", "./web/static")
//	router.LoadHTMLGlob("web/templates/*")
//
//	// Dashboard routes
//	router.GET("/", func(c *gin.Context) {
//		c.HTML(http.StatusOK, "dashboard.html", gin.H{
//			"title": "Real-Time Financial Analysis Dashboard",
//		})
//	})
//
//	// API routes
//	api := router.Group("/api/v1")
//	{
//		api.GET("/dashboard-data", func(c *gin.Context) {
//			data := analysisService.GetDashboardData(c.Request.Context())
//			c.JSON(http.StatusOK, data)
//		})
//
//		api.GET("/sentiment-trends", func(c *gin.Context) {
//			trends := analysisService.GetSentimentTrends(c.Request.Context(), 24*time.Hour)
//			c.JSON(http.StatusOK, trends)
//		})
//
//		api.GET("/impact-predictions", func(c *gin.Context) {
//			predictions := analysisService.GetHighImpactPredictions(c.Request.Context())
//			c.JSON(http.StatusOK, predictions)
//		})
//
//		api.GET("/market-data", func(c *gin.Context) {
//			marketData := analysisService.GetCurrentMarketData(c.Request.Context())
//			c.JSON(http.StatusOK, marketData)
//		})
//
//		api.POST("/trigger-analysis", func(c *gin.Context) {
//			go func() {
//				if err := analysisService.TriggerManualAnalysis(context.Background()); err != nil {
//					logrus.WithError(err).Error("Manual analysis failed")
//				}
//			}()
//			c.JSON(http.StatusOK, gin.H{"message": "Analysis triggered"})
//		})
//	}
//
//	// WebSocket endpoint
//	upgrader := websocket.Upgrader{
//		CheckOrigin: func(r *http.Request) bool {
//			return true // Allow all origins in development
//		},
//	}
//
//	router.GET("/ws", func(c *gin.Context) {
//		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
//		if err != nil {
//			logrus.WithError(err).Error("WebSocket upgrade failed")
//			return
//		}
//
//		client := &analysis.WebSocketClient{
//			Hub:  wsHub,
//			Conn: conn,
//			Send: make(chan []byte, 256),
//		}
//
//		wsHub.Register <- client
//
//		go client.WritePump()
//		go client.ReadPump()
//	})
//
//	return router
//}
//
//func initDatabase() (*database.Database, error) {
//	dbConfig := database.Config{
//		Host:     viper.GetString("database.postgres.host"),
//		Port:     viper.GetInt("database.postgres.port"),
//		Database: viper.GetString("database.postgres.database"),
//		Username: viper.GetString("database.postgres.username"),
//		Password: viper.GetString("database.postgres.password"),
//		SSLMode:  viper.GetString("database.postgres.ssl_mode"),
//	}
//
//	return database.NewConnection(dbConfig)
//}
//
//func initConfig() error {
//	viper.SetConfigName("config")
//	viper.SetConfigType("yaml")
//	viper.AddConfigPath("./config")
//	viper.AddConfigPath(".")
//
//	// Set defaults
//	viper.SetDefault("analysis.dashboard_port", 8080)
//	viper.SetDefault("analysis.update_interval", "30s")
//	viper.SetDefault("analysis.sentiment_threshold", 0.1)
//	viper.SetDefault("analysis.impact_threshold", 1.5)
//	viper.SetDefault("market_data.alpha_vantage_key", "")
//	viper.SetDefault("market_data.yahoo_finance_enabled", true)
//
//	if err := viper.ReadInConfig(); err != nil {
//		logrus.Warnf("Config file not found, using defaults: %v", err)
//	}
//
//	return nil
//}
//
//func initLogger() {
//	logrus.SetLevel(logrus.InfoLevel)
//	logrus.SetFormatter(&logrus.TextFormatter{
//		DisableColors: false,
//		FullTimestamp: true,
//	})
//}
