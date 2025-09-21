// services/news-ingestion/internal/repository/article_repository.go
package repository

import (
	"context"
	_ "fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
)

type ArticleRepository interface {
	Create(ctx context.Context, article *model.Article) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Article, error)
	GetByContentHash(ctx context.Context, hash string) (*model.Article, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.ProcessingStatus) error
	GetPendingArticles(ctx context.Context, limit int) ([]*model.Article, error)
	GetBySymbols(ctx context.Context, symbols []string, limit int) ([]*model.Article, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*model.Article, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type articleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &articleRepository{db: db}
}

func (r *articleRepository) Create(ctx context.Context, article *model.Article) error {
	return r.db.WithContext(ctx).Create(article).Error
}

func (r *articleRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Article, error) {
	var article model.Article
	err := r.db.WithContext(ctx).
		Preload("Source").
		First(&article, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *articleRepository) GetByContentHash(ctx context.Context, hash string) (*model.Article, error) {
	var article model.Article
	err := r.db.WithContext(ctx).
		First(&article, "content_hash = ?", hash).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *articleRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ProcessingStatus) error {
	return r.db.WithContext(ctx).
		Model(&model.Article{}).
		Where("id = ?", id).
		Update("processing_status", status).Error
}

func (r *articleRepository) GetPendingArticles(ctx context.Context, limit int) ([]*model.Article, error) {
	var articles []*model.Article
	err := r.db.WithContext(ctx).
		Where("processing_status = ?", model.StatusPending).
		Order("published_at ASC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func (r *articleRepository) GetBySymbols(ctx context.Context, symbols []string, limit int) ([]*model.Article, error) {
	var articles []*model.Article
	err := r.db.WithContext(ctx).
		Where("symbols && ?", symbols).
		Order("published_at DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func (r *articleRepository) GetByDateRange(ctx context.Context, start, end time.Time) ([]*model.Article, error) {
	var articles []*model.Article
	err := r.db.WithContext(ctx).
		Where("published_at BETWEEN ? AND ?", start, end).
		Order("published_at DESC").
		Find(&articles).Error
	return articles, err
}

func (r *articleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Article{}, "id = ?", id).Error
}

// services/news-ingestion/internal/repository/source_repository.go
package repository

import (
"context"

"gorm.io/gorm"

"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
)

type SourceRepository interface {
	Create(ctx context.Context, source *model.NewsSource) error
	GetByID(ctx context.Context, id uint) (*model.NewsSource, error)
	GetByName(ctx context.Context, name string) (*model.NewsSource, error)
	GetActive(ctx context.Context) ([]*model.NewsSource, error)
	Update(ctx context.Context, source *model.NewsSource) error
	UpdateSuccessRate(ctx context.Context, id uint, rate float64) error
	Delete(ctx context.Context, id uint) error
}

type sourceRepository struct {
	db *gorm.DB
}

func NewSourceRepository(db *gorm.DB) SourceRepository {
	return &sourceRepository{db: db}
}

func (r *sourceRepository) Create(ctx context.Context, source *model.NewsSource) error {
	return r.db.WithContext(ctx).Create(source).Error
}

func (r *sourceRepository) GetByID(ctx context.Context, id uint) (*model.NewsSource, error) {
	var source model.NewsSource
	err := r.db.WithContext(ctx).First(&source, id).Error
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *sourceRepository) GetByName(ctx context.Context, name string) (*model.NewsSource, error) {
	var source model.NewsSource
	err := r.db.WithContext(ctx).First(&source, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *sourceRepository) GetActive(ctx context.Context) ([]*model.NewsSource, error) {
	var sources []*model.NewsSource
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Find(&sources).Error
	return sources, err
}

func (r *sourceRepository) Update(ctx context.Context, source *model.NewsSource) error {
	return r.db.WithContext(ctx).Save(source).Error
}

func (r *sourceRepository) UpdateSuccessRate(ctx context.Context, id uint, rate float64) error {
	return r.db.WithContext(ctx).
		Model(&model.NewsSource{}).
		Where("id = ?", id).
		Update("success_rate", rate).Error
}

func (r *sourceRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.NewsSource{}, id).Error
}

// services/news-ingestion/internal/repository/processing_log_repository.go
package repository

import (
"context"
"time"

"github.com/google/uuid"
"gorm.io/gorm"

"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
)

type ProcessingLogRepository interface {
	Create(ctx context.Context, log *model.ArticleProcessingLog) error
	GetByArticleID(ctx context.Context, articleID uuid.UUID) ([]*model.ArticleProcessingLog, error)
	UpdateProcessingTime(ctx context.Context, id uint, timeMs int) error
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

func (r *processingLogRepository) UpdateProcessingTime(ctx context.Context, id uint, timeMs int) error {
	return r.db.WithContext(ctx).
		Model(&model.ArticleProcessingLog{}).
		Where("id = ?", id).
		Update("processing_time_ms", timeMs).Error
}

// services/news-ingestion/internal/repository/rate_limit_repository.go
package repository

import (
"context"
"time"

"gorm.io/gorm"

"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
)

type RateLimitRepository interface {
	IncrementCount(ctx context.Context, sourceID uint, timeWindow time.Time) error
	GetCurrentCount(ctx context.Context, sourceID uint, timeWindow time.Time) (int, error)
	CleanupOldRecords(ctx context.Context, cutoff time.Time) error
}

type rateLimitRepository struct {
	db *gorm.DB
}

func NewRateLimitRepository(db *gorm.DB) RateLimitRepository {
	return &rateLimitRepository{db: db}
}

func (r *rateLimitRepository) IncrementCount(ctx context.Context, sourceID uint, timeWindow time.Time) error {
	// Use ON CONFLICT for PostgreSQL upsert
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO rate_limit_tracking (source_id, time_window, request_count, created_at, updated_at)
		VALUES (?, ?, 1, NOW(), NOW())
		ON CONFLICT (source_id, time_window)
		DO UPDATE SET
			request_count = rate_limit_tracking.request_count + 1,
			updated_at = NOW()
	`, sourceID, timeWindow).Error
}

func (r *rateLimitRepository) GetCurrentCount(ctx context.Context, sourceID uint, timeWindow time.Time) (int, error) {
	var tracking model.RateLimitTracking
	err := r.db.WithContext(ctx).
		Where("source_id = ? AND time_window = ?", sourceID, timeWindow).
		First(&tracking).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return tracking.RequestCount, nil
}

func (r *rateLimitRepository) CleanupOldRecords(ctx context.Context, cutoff time.Time) error {
	return r.db.WithContext(ctx).
		Where("time_window < ?", cutoff).
		Delete(&model.RateLimitTracking{}).Error
}