// vices/news-ingestion/internal/repository/rate_limit_repository.go
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
