// services/news-ingestion/internal/model/source.go
package model

import (
	"gorm.io/gorm"
	"time"
)

type NewsSource struct {
	ID                 uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name               string         `gorm:"uniqueIndex;not null" json:"name"`
	SourceType         string         `gorm:"not null" json:"source_type"` // RSS, API, TWITTER
	BaseURL            string         `gorm:"type:text" json:"base_url"`
	RateLimitPerMinute int            `gorm:"default:60" json:"rate_limit_per_minute"`
	Status             string         `gorm:"default:active" json:"status"` // active, inactive, error
	SuccessRate        float64        `gorm:"type:decimal(5,2);default:0.00" json:"success_rate"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}
