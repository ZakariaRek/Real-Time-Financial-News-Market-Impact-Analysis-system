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
