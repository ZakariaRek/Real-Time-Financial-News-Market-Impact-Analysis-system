package database

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
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
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable" // For local development
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, sslMode,
	)

	// Configure GORM logger based on environment
	var gormLogger logger.Interface
	if viper.GetString("server.environment") == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Open database connection with improved config
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: true, // This helps with migration issues
		PrepareStmt:                              true, // Improves performance
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
		maxOpenConns = 25 // Good default for AWS RDS
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)

	maxIdleConns := cfg.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 5
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)

	// Parse lifetime durations
	if cfg.ConnMaxLifetime != "" {
		if lifetime, err := time.ParseDuration(cfg.ConnMaxLifetime); err == nil {
			sqlDB.SetConnMaxLifetime(lifetime)
		}
	} else {
		sqlDB.SetConnMaxLifetime(5 * time.Minute) // Default for AWS RDS
	}

	if cfg.ConnMaxIdleTime != "" {
		if idleTime, err := time.ParseDuration(cfg.ConnMaxIdleTime); err == nil {
			sqlDB.SetConnMaxIdleTime(idleTime)
		}
	} else {
		sqlDB.SetConnMaxIdleTime(1 * time.Minute)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.Info("Successfully connected to PostgreSQL database")

	return &Database{DB: db}, nil
}

func (d *Database) AutoMigrate() error {
	// Enable UUID extension first
	if err := d.enableUUIDExtension(); err != nil {
		logrus.WithError(err).Warn("Failed to enable UUID extension, continuing anyway")
	}

	// Migrate in the correct order to handle foreign key dependencies
	logrus.Info("Starting database migrations...")

	// First migrate tables without foreign keys
	err := d.DB.AutoMigrate(&model.NewsSource{})
	if err != nil {
		return fmt.Errorf("failed to migrate NewsSource: %w", err)
	}
	logrus.Info("NewsSource table migrated successfully")

	// Then migrate tables with foreign keys
	err = d.DB.AutoMigrate(&model.Article{})
	if err != nil {
		return fmt.Errorf("failed to migrate Article: %w", err)
	}
	logrus.Info("Article table migrated successfully")

	err = d.DB.AutoMigrate(&model.ArticleProcessingLog{})
	if err != nil {
		return fmt.Errorf("failed to migrate ArticleProcessingLog: %w", err)
	}
	logrus.Info("ArticleProcessingLog table migrated successfully")

	err = d.DB.AutoMigrate(&model.RateLimitTracking{})
	if err != nil {
		return fmt.Errorf("failed to migrate RateLimitTracking: %w", err)
	}
	logrus.Info("RateLimitTracking table migrated successfully")

	logrus.Info("Database migrations completed successfully")
	return nil
}

func (d *Database) enableUUIDExtension() error {
	// Try to enable uuid-ossp extension first
	err := d.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	if err != nil {
		logrus.WithError(err).Debug("Failed to enable uuid-ossp extension, trying pgcrypto")
	}

	// Also try pgcrypto for gen_random_uuid() if uuid-ossp is not available
	err = d.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"").Error
	if err != nil {
		logrus.WithError(err).Debug("Failed to enable pgcrypto extension")
		return fmt.Errorf("failed to enable UUID extensions: %w", err)
	}

	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateIndexes creates additional indexes for performance optimization
func (d *Database) CreateIndexes() error {
	// Create composite indexes for better query performance
	indexes := []string{
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_status_published ON articles(processing_status, published_at)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_articles_symbols_gin ON articles USING GIN(symbols)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rate_limit_source_time ON rate_limit_tracking(source_id, time_window)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_processing_logs_stage_status ON article_processing_logs(processing_stage, status)",
		"CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_rate_limit_source_window ON rate_limit_tracking(source_id, time_window)",
	}

	for _, index := range indexes {
		if err := d.DB.Exec(index).Error; err != nil {
			logrus.WithError(err).Warnf("Failed to create index: %s", index)
			// Don't return error for index creation failures in case they already exist
		}
	}

	return nil
}
