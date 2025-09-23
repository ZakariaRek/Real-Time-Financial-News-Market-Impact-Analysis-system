package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
	"github.com/redis/go-redis/v9"
)

type CacheRepository interface {
	SetAnalysisResult(ctx context.Context, articleID string, result *model.AnalysisResult, ttl time.Duration) error
	GetAnalysisResult(ctx context.Context, articleID string) (*model.AnalysisResult, error)
	SetModelPrediction(ctx context.Context, key string, prediction interface{}, ttl time.Duration) error
	GetModelPrediction(ctx context.Context, key string, dest interface{}) error
	DeleteAnalysisResult(ctx context.Context, articleID string) error
	SetProcessingStatus(ctx context.Context, articleID string, status string, ttl time.Duration) error
	GetProcessingStatus(ctx context.Context, articleID string) (string, error)
}

type cacheRepository struct {
	redis *redis.Client
}

func NewCacheRepository(redis *redis.Client) CacheRepository {
	return &cacheRepository{redis: redis}
}

func (r *cacheRepository) SetAnalysisResult(ctx context.Context, articleID string, result *model.AnalysisResult, ttl time.Duration) error {
	key := fmt.Sprintf("analysis:result:%s", articleID)
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return r.redis.Set(ctx, key, data, ttl).Err()
}

func (r *cacheRepository) GetAnalysisResult(ctx context.Context, articleID string) (*model.AnalysisResult, error) {
	key := fmt.Sprintf("analysis:result:%s", articleID)
	data, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result model.AnalysisResult
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *cacheRepository) SetModelPrediction(ctx context.Context, key string, prediction interface{}, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("model:prediction:%s", key)
	data, err := json.Marshal(prediction)
	if err != nil {
		return err
	}
	return r.redis.Set(ctx, cacheKey, data, ttl).Err()
}

func (r *cacheRepository) GetModelPrediction(ctx context.Context, key string, dest interface{}) error {
	cacheKey := fmt.Sprintf("model:prediction:%s", key)
	data, err := r.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

func (r *cacheRepository) DeleteAnalysisResult(ctx context.Context, articleID string) error {
	key := fmt.Sprintf("analysis:result:%s", articleID)
	return r.redis.Del(ctx, key).Err()
}

func (r *cacheRepository) SetProcessingStatus(ctx context.Context, articleID string, status string, ttl time.Duration) error {
	key := fmt.Sprintf("analysis:status:%s", articleID)
	return r.redis.Set(ctx, key, status, ttl).Err()
}

func (r *cacheRepository) GetProcessingStatus(ctx context.Context, articleID string) (string, error) {
	key := fmt.Sprintf("analysis:status:%s", articleID)
	return r.redis.Get(ctx, key).Result()
}
