// services/nlp-processing/internal/model/entity.go
package model

import (
	"time"

	"gorm.io/gorm"
)

type EntityRecognition struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleID      string         `gorm:"type:varchar(255);not null;index" json:"article_id"`
	EntityText     string         `gorm:"type:text;not null" json:"entity_text"`
	EntityType     string         `gorm:"type:varchar(50);not null" json:"entity_type"` // PERSON, ORG, MONEY, etc.
	StockSymbol    string         `gorm:"type:varchar(20);index" json:"stock_symbol"`
	Confidence     float32        `gorm:"type:real;not null" json:"confidence"`
	EntityCategory string         `gorm:"type:varchar(50)" json:"entity_category"` // Company, Executive, Financial_Instrument
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
