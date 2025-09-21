// services/nlp-processing/internal/database/connection.go
package database

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
)

type Config struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Database        string `mapstructure:"database"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime string `mapstructure:"conn_max_idle_time"`
}

type Database struct {
	DB *gorm.DB
}

func NewConnection(cfg Config) (*Database, error) {
	// Build PostgreSQL DSN
	// Note: For high-volume analytics in production, consider using ClickHouse on AWS
	// This PostgreSQL implementation is for compatibility with the requested architecture
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable" // For local development
		// In production with AWS RDS, use: sslMode = "require"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, sslMode,
	)

	// Configure GORM logger based on environment
	var gormLogger logger.Interface
	if viper.GetString("server.environment") == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool for AWS RDS optimization
	maxOpenConns := cfg.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 30 // Higher for analytics workload
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)

	maxIdleConns := cfg.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 10
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)

	// Parse lifetime durations
	if cfg.ConnMaxLifetime != "" {
		if lifetime, err := time.ParseDuration(cfg.ConnMaxLifetime); err == nil {
			sqlDB.SetConnMaxLifetime(lifetime)
		}
	} else {
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}

	if cfg.ConnMaxIdleTime != "" {
		if idleTime, err := time.ParseDuration(cfg.ConnMaxIdleTime); err == nil {
			sqlDB.SetConnMaxIdleTime(idleTime)
		}
	} else {
		sqlDB.SetConnMaxIdleTime(2 * time.Minute)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.Info("Successfully connected to NLP Processing PostgreSQL database")

	return &Database{DB: db}, nil
}

func (d *Database) AutoMigrate() error {
	// Run database migrations
	// In production with AWS RDS, consider using golang-migrate for better control
	err := d.DB.AutoMigrate(
		&model.SentimentAnalysis{},
		&model.EntityRecognition{},
		&model.TopicClassification{},
		&model.SentimentHourlyMV{},
	)
	if err != nil {
		return fmt.Errorf("failed to run auto migrations: %w", err)
	}

	logrus.Info("NLP Processing database migrations completed successfully")
	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateIndexes creates additional indexes for analytics performance
func (d *Database) CreateIndexes() error {
	// Create composite indexes for high-performance analytics queries
	// Note: For production analytics at scale, consider ClickHouse on AWS
	indexes := []string{
		// Sentiment analysis indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sentiment_symbol_timestamp ON sentiment_analysis(primary_symbol, analysis_timestamp DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sentiment_compound_score ON sentiment_analysis(compound_score) WHERE compound_score IS NOT NULL",

		// Entity recognition indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_entity_type_symbol ON entity_recognition(entity_type, stock_symbol)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_entity_confidence ON entity_recognition(confidence) WHERE confidence > 0.7",

		// Topic classification indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_topic_primary_urgency ON topic_classification(primary_topic, urgency_score DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_topic_breaking_news ON topic_classification(breaking_news_indicator) WHERE breaking_news_indicator = 1",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_topic_keywords_gin ON topic_classification USING GIN(keywords)",

		// Sentiment hourly materialized view indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sentiment_hourly_symbol_time ON sentiment_hourly_mv(primary_symbol, hour_timestamp DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sentiment_hourly_volatility ON sentiment_hourly_mv(sentiment_volatility DESC)",
	}

	for _, index := range indexes {
		if err := d.DB.Exec(index).Error; err != nil {
			logrus.WithError(err).Warnf("Failed to create index: %s", index)
			// Don't return error for index creation failures in case they already exist
		}
	}

	return nil
}

// CreateMaterializedViews creates materialized views for analytics
func (d *Database) CreateMaterializedViews() error {
	// Create materialized view for hourly sentiment aggregation
	// This provides fast analytics queries for dashboard and reporting
	createMVSQL := `
		CREATE MATERIALIZED VIEW IF NOT EXISTS sentiment_hourly_mv AS
		SELECT 
			date_trunc('hour', analysis_timestamp) as hour_timestamp,
			primary_symbol,
			COUNT(*)::bigint as article_count,
			AVG(compound_score)::double precision as avg_sentiment,
			STDDEV(compound_score)::double precision as sentiment_volatility,
			NOW() as created_at,
			NOW() as updated_at
		FROM sentiment_analysis 
		WHERE primary_symbol IS NOT NULL 
		  AND analysis_timestamp >= NOW() - INTERVAL '30 days'
		GROUP BY date_trunc('hour', analysis_timestamp), primary_symbol
		ORDER BY hour_timestamp DESC, primary_symbol;
	`

	if err := d.DB.Exec(createMVSQL).Error; err != nil {
		logrus.WithError(err).Warn("Failed to create sentiment hourly materialized view")
	}

	// Create index on materialized view
	mvIndexSQL := `
		CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_sentiment_hourly_mv_pk 
		ON sentiment_hourly_mv(hour_timestamp, primary_symbol);
	`

	if err := d.DB.Exec(mvIndexSQL).Error; err != nil {
		logrus.WithError(err).Warn("Failed to create materialized view index")
	}

	// Note: In production, set up a cron job or scheduled task to refresh this materialized view
	// Example: REFRESH MATERIALIZED VIEW CONCURRENTLY sentiment_hourly_mv;

	return nil
}

// RefreshMaterializedViews refreshes the materialized views for up-to-date analytics
func (d *Database) RefreshMaterializedViews() error {
	refreshSQL := "REFRESH MATERIALIZED VIEW CONCURRENTLY sentiment_hourly_mv"

	if err := d.DB.Exec(refreshSQL).Error; err != nil {
		return fmt.Errorf("failed to refresh materialized views: %w", err)
	}

	logrus.Info("Materialized views refreshed successfully")
	return nil
}
