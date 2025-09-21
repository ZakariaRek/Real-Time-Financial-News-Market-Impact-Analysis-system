// // services/news-ingestion/internal/model/models.go
package model

//
//import (
//	"time"
//
//	"github.com/google/uuid"
//	"github.com/lib/pq"
//	"gorm.io/gorm"
//)
//
//type ProcessingStatus string
//
//const (
//	StatusPending    ProcessingStatus = "pending"
//	StatusProcessing ProcessingStatus = "processing"
//	StatusCompleted  ProcessingStatus = "completed"
//	StatusFailed     ProcessingStatus = "failed"
//	StatusSkipped    ProcessingStatus = "skipped"
//)
//
//// NewsSource represents a news source configuration
//type NewsSource struct {
//	ID            uint      `json:"id" gorm:"primaryKey"`
//	Name          string    `json:"name" gorm:"uniqueIndex;not null"`
//	Type          string    `json:"type" gorm:"not null"` // rss, api, twitter
//	URL           string    `json:"url"`
//	APIKey        string    `json:"api_key,omitempty"`
//	RateLimit     int       `json:"rate_limit" gorm:"default:60"` // requests per hour
//	Status        string    `json:"status" gorm:"default:active"`
//	SuccessRate   float64   `json:"success_rate" gorm:"default:0.0"`
//	LastFetchedAt time.Time `json:"last_fetched_at"`
//	CreatedAt     time.Time `json:"created_at"`
//	UpdatedAt     time.Time `json:"updated_at"`
//}
//
//// Article represents a news article
//type Article struct {
//	ID               uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
//	SourceID         uint             `json:"source_id" gorm:"not null"`
//	Source           NewsSource       `json:"source" gorm:"foreignKey:SourceID"`
//	Title            string           `json:"title" gorm:"not null"`
//	Content          string           `json:"content" gorm:"type:text"`
//	Summary          string           `json:"summary" gorm:"type:text"`
//	Author           string           `json:"author"`
//	URL              string           `json:"url" gorm:"uniqueIndex"`
//	ImageURL         string           `json:"image_url"`
//	PublishedAt      time.Time        `json:"published_at"`
//	ContentHash      string           `json:"content_hash" gorm:"uniqueIndex;not null"`
//	ProcessingStatus ProcessingStatus `json:"processing_status" gorm:"default:pending"`
//	Symbols          pq.StringArray   `json:"symbols" gorm:"type:text[]"`
//	Sentiment        float64          `json:"sentiment" gorm:"default:0.0"`
//	Impact           string           `json:"impact"`
//	Keywords         pq.StringArray   `json:"keywords" gorm:"type:text[]"`
//	Category         string           `json:"category"`
//	Language         string           `json:"language" gorm:"default:en"`
//	CreatedAt        time.Time        `json:"created_at"`
//	UpdatedAt        time.Time        `json:"updated_at"`
//}
//
//// ArticleProcessingLog tracks processing steps for articles
//type ArticleProcessingLog struct {
//	ID               uint      `json:"id" gorm:"primaryKey"`
//	ArticleID        uuid.UUID `json:"article_id" gorm:"type:uuid;not null"`
//	Article          Article   `json:"article" gorm:"foreignKey:ArticleID"`
//	ProcessingStage  string    `json:"processing_stage" gorm:"not null"`
//	Status           string    `json:"status" gorm:"not null"`
//	Message          string    `json:"message"`
//	ProcessingTimeMs int       `json:"processing_time_ms"`
//	ErrorDetails     string    `json:"error_details,omitempty"`
//	CreatedAt        time.Time `json:"created_at"`
//}
//
//// RateLimitTracking tracks API usage for rate limiting
//type RateLimitTracking struct {
//	ID           uint       `json:"id" gorm:"primaryKey"`
//	SourceID     uint       `json:"source_id" gorm:"not null"`
//	Source       NewsSource `json:"source" gorm:"foreignKey:SourceID"`
//	TimeWindow   time.Time  `json:"time_window" gorm:"not null"`
//	RequestCount int        `json:"request_count" gorm:"default:0"`
//	CreatedAt    time.Time  `json:"created_at"`
//	UpdatedAt    time.Time  `json:"updated_at"`
//}
//
//// Add unique constraint on source_id and time_window
//func (RateLimitTracking) TableName() string {
//	return "rate_limit_tracking"
//}
//
//// BeforeCreate hooks
//func (a *Article) BeforeCreate(tx *gorm.DB) (err error) {
//	if a.ID == uuid.Nil {
//		a.ID = uuid.New()
//	}
//	return
//}
