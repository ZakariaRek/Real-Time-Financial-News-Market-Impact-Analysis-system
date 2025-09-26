// services/nlp-processing/internal/repository/processing_log_repository.go
package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
)

type ProcessingLogRepository interface {
	Create(ctx context.Context, log *model.ArticleProcessingLog) error
	GetByArticleID(ctx context.Context, articleID uuid.UUID) ([]*model.ArticleProcessingLog, error)
	GetByStatus(ctx context.Context, status string, limit int) ([]*model.ArticleProcessingLog, error)
}

type processingLogRepository struct {
	db *gorm.DB
}

func NewProcessingLogRepository(db *gorm.DB) ProcessingLogRepository {
	return &processingLogRepository{db: db}
}

func (r *processingLogRepository) Create(ctx context.Context, log *model.ArticleProcessingLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *processingLogRepository) GetByArticleID(ctx context.Context, articleID uuid.UUID) ([]*model.ArticleProcessingLog, error) {
	var logs []*model.ArticleProcessingLog
	err := r.db.WithContext(ctx).
		Where("article_id = ?", articleID).
		Order("created_at ASC").
		Find(&logs).Error
	return logs, err
}

func (r *processingLogRepository) GetByStatus(ctx context.Context, status string, limit int) ([]*model.ArticleProcessingLog, error) {
	var logs []*model.ArticleProcessingLog
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
