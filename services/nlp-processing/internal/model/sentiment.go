// services/nlp-processing/internal/model/sentiment.go
package model

import (
	"time"
)

// Materialized view for hourly sentiment aggregation
// Note: PostgreSQL implementation - for production analytics consider ClickHouse
type SentimentHourlyMV struct {
	HourTimestamp       time.Time `gorm:"type:timestamp;primaryKey" json:"hour_timestamp"`
	PrimarySymbol       string    `gorm:"type:varchar(50);primaryKey" json:"primary_symbol"`
	ArticleCount        uint64    `gorm:"type:bigint;not null" json:"article_count"`
	AvgSentiment        float64   `gorm:"type:double precision;not null" json:"avg_sentiment"`
	SentimentVolatility float64   `gorm:"type:double precision;not null" json:"sentiment_volatility"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Model for comprehensive analysis result
type AnalysisResult struct {
	ArticleID           string               `json:"article_id"`
	SentimentAnalysis   *SentimentAnalysis   `json:"sentiment_analysis,omitempty"`
	EntityRecognition   []EntityRecognition  `json:"entity_recognition,omitempty"`
	TopicClassification *TopicClassification `json:"topic_classification,omitempty"`
	ProcessingTimeMs    int64                `json:"processing_time_ms"`
	Status              string               `json:"status"`
	ErrorMessage        string               `json:"error_message,omitempty"`
}

// Request/Response models for API
type AnalysisRequest struct {
	ArticleID string   `json:"article_id" binding:"required"`
	Title     string   `json:"title" binding:"required"`
	Content   string   `json:"content" binding:"required"`
	URL       string   `json:"url"`
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
	Timestamp    time.Time `json:"timestamp"`
	Symbol       string    `json:"symbol"`
	AvgSentiment float64   `json:"avg_sentiment"`
	ArticleCount int64     `json:"article_count"`
	Volatility   float64   `json:"volatility"`
}
