// services/news-ingestion/internal/model/source.go
package model

import (
	"time"

	"gorm.io/gorm"
)

type NewsSource struct {
	ID                 uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name               string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	SourceType         string         `gorm:"type:varchar(50);not null" json:"source_type"` // RSS, API, TWITTER
	BaseURL            string         `gorm:"type:text" json:"base_url"`
	RateLimitPerMinute int            `gorm:"default:60" json:"rate_limit_per_minute"`
	Status             string         `gorm:"type:varchar(50);default:'active'" json:"status"` // active, inactive, error
	SuccessRate        float64        `gorm:"type:decimal(5,2);default:0.00" json:"success_rate"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for NewsSource
func (NewsSource) TableName() string {
	return "news_sources"
}

// BeforeCreate hook
func (ns *NewsSource) BeforeCreate(tx *gorm.DB) (err error) {
	// Ensure status is set
	if ns.Status == "" {
		ns.Status = "active"
	}

	// Ensure rate limit is set
	if ns.RateLimitPerMinute <= 0 {
		ns.RateLimitPerMinute = 60
	}

	return
}
