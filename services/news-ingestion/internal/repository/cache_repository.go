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
