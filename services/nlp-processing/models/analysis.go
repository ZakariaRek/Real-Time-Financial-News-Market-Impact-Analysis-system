// services/nlp-processing/internal/model/analysis.go
package model

import (
	"time"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Note: Originally designed for ClickHouse for high-performance analytics
// Implemented with PostgreSQL as requested, but consider ClickHouse for production
// with high-volume financial data analytics on AWS

type SentimentAnalysis struct {
	ArticleID         string    `gorm:"type:varchar(255);primaryKey" json:"article_id"`
	AnalysisTimestamp time.Time `gorm:"type:timestamp;not null;index" json:"analysis_timestamp"`
	CompoundScore     float32   `gorm:"type:real;not null" json:"compound_score"`
	Confidence        float32   `gorm:"type:real;not null" json:"confidence"`
	PrimarySymbol     string    `gorm:"type:varchar(50);index" json:"primary_symbol"`
	ModelVersion      string    `gorm:"type:varchar(50);not null" json:"model_version"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// services/nlp-processing/internal/model/entity.go
package model

import (
"time"
"gorm.io/gorm"
)

type EntityRecognition struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleID      string    `gorm:"type:varchar(255);not null;index" json:"article_id"`
	EntityText     string    `gorm:"type:text;not null" json:"entity_text"`
	EntityType     string    `gorm:"type:varchar(50);not null" json:"entity_type"` // PERSON, ORG, MONEY, etc.
	StockSymbol    string    `gorm:"type:varchar(20);index" json:"stock_symbol"`
	Confidence     float32   `gorm:"type:real;not null" json:"confidence"`
	EntityCategory string    `gorm:"type:varchar(50)" json:"entity_category"` // Company, Executive, Financial_Instrument
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// services/nlp-processing/internal/model/topic.go
package model

import (
"time"
"github.com/lib/pq"
"gorm.io/gorm"
)

type TopicClassification struct {
	ID                      uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleID               string         `gorm:"type:varchar(255);not null;index" json:"article_id"`
	PrimaryTopic            string         `gorm:"type:varchar(100);not null;index" json:"primary_topic"`
	PrimaryTopicConfidence  float32        `gorm:"type:real;not null" json:"primary_topic_confidence"`
	Keywords                pq.StringArray `gorm:"type:text[]" json:"keywords"`
	UrgencyScore            float32        `gorm:"type:real;default:0.0" json:"urgency_score"`
	BreakingNewsIndicator   uint8          `gorm:"type:smallint;default:0" json:"breaking_news_indicator"` // 0 or 1
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"-"`
}

// services/nlp-processing/internal/model/sentiment.go
package model

import (
"time"
)

// Materialized view for hourly sentiment aggregation
// Note: PostgreSQL implementation - for production analytics consider ClickHouse
type SentimentHourlyMV struct {
	HourTimestamp      time.Time `gorm:"type:timestamp;primaryKey" json:"hour_timestamp"`
	PrimarySymbol      string    `gorm:"type:varchar(50);primaryKey" json:"primary_symbol"`
	ArticleCount       uint64    `gorm:"type:bigint;not null" json:"article_count"`
	AvgSentiment       float64   `gorm:"type:double precision;not null" json:"avg_sentiment"`
	SentimentVolatility float64  `gorm:"type:double precision;not null" json:"sentiment_volatility"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Model for comprehensive analysis result
type AnalysisResult struct {
	ArticleID            string              `json:"article_id"`
	SentimentAnalysis    *SentimentAnalysis  `json:"sentiment_analysis,omitempty"`
	EntityRecognition    []EntityRecognition `json:"entity_recognition,omitempty"`
	TopicClassification  *TopicClassification `json:"topic_classification,omitempty"`
	ProcessingTimeMs     int64               `json:"processing_time_ms"`
	Status               string              `json:"status"`
	ErrorMessage         string              `json:"error_message,omitempty"`
}

// Request/Response models for API
type AnalysisRequest struct {
	ArticleID string `json:"article_id" binding:"required"`
	Title     string `json:"title" binding:"required"`
	Content   string `json:"content" binding:"required"`
	URL       string `json:"url"`
	Symbols   []string `json:"symbols"`
}

type BatchAnalysisRequest struct {
	Articles []AnalysisRequest `json:"articles" binding:"required,min=1,max=100"`
}

type AnalysisResponse struct {
	Success bool             `json:"success"`
	Results []AnalysisResult `json:"results"`
	Errors  []string         `json:"errors,omitempty"`
}

// Analytics query models
type SentimentQuery struct {
	Symbol    string    `json:"symbol"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Interval  string    `json:"interval"` // hour, day, week
}

type SentimentTrend struct {
	Timestamp   time.Time `json:"timestamp"`
	Symbol      string    `json:"symbol"`
	AvgSentiment float64  `json:"avg_sentiment"`
	ArticleCount int64    `json:"article_count"`
	Volatility   float64  `json:"volatility"`
}