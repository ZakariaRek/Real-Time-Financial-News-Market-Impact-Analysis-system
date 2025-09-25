// services/nlp-processing/internal/model/request_response.go
package model

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Processing status enum
type ProcessingStatus string

const (
	StatusPending    ProcessingStatus = "pending"
	StatusProcessing ProcessingStatus = "processing"
	StatusCompleted  ProcessingStatus = "completed"
	StatusFailed     ProcessingStatus = "failed"
)

// Implement driver.Valuer interface for ProcessingStatus
func (ps ProcessingStatus) Value() (driver.Value, error) {
	return string(ps), nil
}

// Implement sql.Scanner interface for ProcessingStatus
func (ps *ProcessingStatus) Scan(value interface{}) error {
	if value == nil {
		*ps = StatusPending
		return nil
	}

	switch s := value.(type) {
	case string:
		*ps = ProcessingStatus(s)
	case []byte:
		*ps = ProcessingStatus(string(s))
	default:
		return errors.New("cannot scan value into ProcessingStatus")
	}
	return nil
}

// Article model for NLP processing
type Article struct {
	ID               uuid.UUID        `gorm:"type:uuid;primaryKey;" json:"id"`
	SourceID         uint             `gorm:"not null;index" json:"source_id"`
	Source           NewsSource       `gorm:"foreignKey:SourceID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"source,omitempty"`
	Title            string           `gorm:"type:text;not null" json:"title"`
	Content          string           `gorm:"type:text" json:"content"`
	URL              string           `gorm:"type:text;not null" json:"url"`
	Symbols          pq.StringArray   `gorm:"type:text[]" json:"symbols"`
	PublishedAt      time.Time        `gorm:"not null;index" json:"published_at"`
	ProcessingStatus ProcessingStatus `gorm:"type:varchar(50);default:'pending';index" json:"processing_status"`
	RelevanceScore   float64          `gorm:"type:decimal(5,4);default:0.0000" json:"relevance_score"`
	ContentHash      string           `gorm:"type:varchar(255);uniqueIndex;not null" json:"content_hash"`
	CreatedAt        time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"-"`
}

// BeforeCreate hook to ensure UUID is generated
func (a *Article) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}

	// Ensure processing status is set
	if a.ProcessingStatus == "" {
		a.ProcessingStatus = StatusPending
	}

	// Ensure published time is set
	if a.PublishedAt.IsZero() {
		a.PublishedAt = time.Now().UTC()
	}

	return
}

// News source model
type NewsSource struct {
	ID                 uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name               string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	SourceType         string         `gorm:"type:varchar(50);not null" json:"source_type"`
	BaseURL            string         `gorm:"type:text" json:"base_url"`
	RateLimitPerMinute int            `gorm:"default:60" json:"rate_limit_per_minute"`
	Status             string         `gorm:"type:varchar(50);default:'active'" json:"status"`
	SuccessRate        float64        `gorm:"type:decimal(5,2);default:0.00" json:"success_rate"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

// Processing log model
type ArticleProcessingLog struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleID        uuid.UUID `gorm:"type:uuid;not null;index" json:"article_id"`
	Article          Article   `gorm:"foreignKey:ArticleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"article,omitempty"`
	ProcessingStage  string    `gorm:"type:varchar(100);not null" json:"processing_stage"`
	Status           string    `gorm:"type:varchar(50);not null" json:"status"`
	ProcessingTimeMs int       `gorm:"default:0" json:"processing_time_ms"`
	ErrorMessage     string    `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Processing options for batch processing
type ProcessingOptions struct {
	EnableSentiment     bool    `json:"enable_sentiment"`
	EnableNER           bool    `json:"enable_ner"`
	EnableTopic         bool    `json:"enable_topic"`
	EnableKeywords      bool    `json:"enable_keywords"`
	ConfidenceThreshold float32 `json:"confidence_threshold"`
	ModelVersion        string  `json:"model_version"`
}

// Model status for health checks
type ModelStatus struct {
	FinbertLoaded    bool   `json:"finbert_loaded"`
	NerModelLoaded   bool   `json:"ner_model_loaded"`
	TopicModelLoaded bool   `json:"topic_model_loaded"`
	FinbertVersion   string `json:"finbert_version"`
	NerVersion       string `json:"ner_version"`
	TopicVersion     string `json:"topic_version"`
}
