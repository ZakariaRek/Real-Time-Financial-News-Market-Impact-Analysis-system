// services/nlp-processing/internal/model/topic.go
package model

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type TopicClassification struct {
	ID                     uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleID              string         `gorm:"type:varchar(255);not null;index" json:"article_id"`
	PrimaryTopic           string         `gorm:"type:varchar(100);not null;index" json:"primary_topic"`
	PrimaryTopicConfidence float32        `gorm:"type:real;not null" json:"primary_topic_confidence"`
	Keywords               pq.StringArray `gorm:"type:text[]" json:"keywords"`
	UrgencyScore           float32        `gorm:"type:real;default:0.0" json:"urgency_score"`
	BreakingNewsIndicator  uint8          `gorm:"type:smallint;default:0" json:"breaking_news_indicator"` // 0 or 1
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`
}
