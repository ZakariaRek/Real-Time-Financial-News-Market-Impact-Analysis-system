// services/news-ingestion/internal/model/source.go
package model

import (
	"time"
	"gorm.io/gorm"
)

type NewsSource struct {
	ID                  uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name                string    `gorm:"uniqueIndex;not null" json:"name"`
	SourceType          string    `gorm:"not null" json:"source_type"` // RSS, API, TWITTER
	BaseURL             string    `gorm:"type:text" json:"base_url"`
	RateLimitPerMinute  int       `gorm:"default:60" json:"rate_limit_per_minute"`
	Status              string    `gorm:"default:active" json:"status"` // active, inactive, error
	SuccessRate         float64   `gorm:"type:decimal(5,2);default:0.00" json:"success_rate"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// services/news-ingestion/internal/model/article.go
package model

import (
"time"
"github.com/google/uuid"
"github.com/lib/pq"
"gorm.io/gorm"
)

type ProcessingStatus string

const (
	StatusPending    ProcessingStatus = "pending"
	StatusProcessing ProcessingStatus = "processing"
	StatusCompleted  ProcessingStatus = "completed"
	StatusFailed     ProcessingStatus = "failed"
)

type Article struct {
	ID               uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SourceID         uint             `gorm:"not null;index" json:"source_id"`
	Source           NewsSource       `gorm:"foreignKey:SourceID" json:"source,omitempty"`
	Title            string           `gorm:"type:text;not null" json:"title"`
	Content          string           `gorm:"type:text" json:"content"`
	URL              string           `gorm:"type:text;not null" json:"url"`
	Symbols          pq.StringArray   `gorm:"type:text[]" json:"symbols"`
	PublishedAt      time.Time        `gorm:"not null;index" json:"published_at"`
	ProcessingStatus ProcessingStatus `gorm:"default:pending;index" json:"processing_status"`
	RelevanceScore   float64          `gorm:"type:decimal(5,4);default:0.0000" json:"relevance_score"`
	ContentHash      string           `gorm:"uniqueIndex;not null" json:"content_hash"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"-"`
}

type ArticleProcessingLog struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleID         uuid.UUID `gorm:"type:uuid;not null;index" json:"article_id"`
	Article           Article   `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
	ProcessingStage   string    `gorm:"not null" json:"processing_stage"` // ingestion, validation, nlp_processing
	Status            string    `gorm:"not null" json:"status"`           // started, completed, failed
	ProcessingTimeMs  int       `gorm:"default:0" json:"processing_time_ms"`
	ErrorMessage      string    `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type RateLimitTracking struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SourceID     uint      `gorm:"not null;index" json:"source_id"`
	Source       NewsSource `gorm:"foreignKey:SourceID" json:"source,omitempty"`
	TimeWindow   time.Time `gorm:"not null;index" json:"time_window"`
	RequestCount int       `gorm:"default:0" json:"request_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}