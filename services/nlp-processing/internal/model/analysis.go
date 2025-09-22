// services/nlp-processing/internal/model/analysis.go
package model

import (
	"time"
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
