// services/nlp-processing/internal/repository/analysis_repository.go
package repository

import (
	"context"
	_ "encoding/json"
	_ "fmt"
	"time"

	"gorm.io/gorm"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
)

type AnalysisRepository interface {
	// Sentiment Analysis
	CreateSentimentAnalysis(ctx context.Context, analysis *model.SentimentAnalysis) error
	GetSentimentByArticleID(ctx context.Context, articleID string) (*model.SentimentAnalysis, error)
	GetSentimentTrends(ctx context.Context, symbol string, start, end time.Time) ([]*model.SentimentAnalysis, error)

	// Entity Recognition
	CreateEntityRecognition(ctx context.Context, entities []model.EntityRecognition) error
	GetEntitiesByArticleID(ctx context.Context, articleID string) ([]*model.EntityRecognition, error)
	GetEntitiesBySymbol(ctx context.Context, symbol string, limit int) ([]*model.EntityRecognition, error)

	// Topic Classification
	CreateTopicClassification(ctx context.Context, topic *model.TopicClassification) error
	GetTopicByArticleID(ctx context.Context, articleID string) (*model.TopicClassification, error)
	GetBreakingNews(ctx context.Context, limit int) ([]*model.TopicClassification, error)

	// Analytics queries
	GetSentimentHourlyData(ctx context.Context, symbol string, hours int) ([]*model.SentimentHourlyMV, error)
	GetTopTopicsLastDay(ctx context.Context, limit int) ([]map[string]interface{}, error)
	GetMostMentionedEntities(ctx context.Context, days int, limit int) ([]map[string]interface{}, error)
}

type analysisRepository struct {
	db *gorm.DB
}

func NewAnalysisRepository(db *gorm.DB) AnalysisRepository {
	return &analysisRepository{db: db}
}

// Sentiment Analysis methods
func (r *analysisRepository) CreateSentimentAnalysis(ctx context.Context, analysis *model.SentimentAnalysis) error {
	return r.db.WithContext(ctx).Create(analysis).Error
}

func (r *analysisRepository) GetSentimentByArticleID(ctx context.Context, articleID string) (*model.SentimentAnalysis, error) {
	var analysis model.SentimentAnalysis
	err := r.db.WithContext(ctx).
		First(&analysis, "article_id = ?", articleID).Error
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

func (r *analysisRepository) GetSentimentTrends(ctx context.Context, symbol string, start, end time.Time) ([]*model.SentimentAnalysis, error) {
	var analyses []*model.SentimentAnalysis
	err := r.db.WithContext(ctx).
		Where("primary_symbol = ? AND analysis_timestamp BETWEEN ? AND ?", symbol, start, end).
		Order("analysis_timestamp ASC").
		Find(&analyses).Error
	return analyses, err
}

// Entity Recognition methods
func (r *analysisRepository) CreateEntityRecognition(ctx context.Context, entities []model.EntityRecognition) error {
	if len(entities) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(entities, 100).Error
}

func (r *analysisRepository) GetEntitiesByArticleID(ctx context.Context, articleID string) ([]*model.EntityRecognition, error) {
	var entities []*model.EntityRecognition
	err := r.db.WithContext(ctx).
		Where("article_id = ?", articleID).
		Order("confidence DESC").
		Find(&entities).Error
	return entities, err
}

func (r *analysisRepository) GetEntitiesBySymbol(ctx context.Context, symbol string, limit int) ([]*model.EntityRecognition, error) {
	var entities []*model.EntityRecognition
	err := r.db.WithContext(ctx).
		Where("stock_symbol = ?", symbol).
		Order("created_at DESC").
		Limit(limit).
		Find(&entities).Error
	return entities, err
}

// Topic Classification methods
func (r *analysisRepository) CreateTopicClassification(ctx context.Context, topic *model.TopicClassification) error {
	return r.db.WithContext(ctx).Create(topic).Error
}

func (r *analysisRepository) GetTopicByArticleID(ctx context.Context, articleID string) (*model.TopicClassification, error) {
	var topic model.TopicClassification
	err := r.db.WithContext(ctx).
		First(&topic, "article_id = ?", articleID).Error
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *analysisRepository) GetBreakingNews(ctx context.Context, limit int) ([]*model.TopicClassification, error) {
	var topics []*model.TopicClassification
	err := r.db.WithContext(ctx).
		Where("breaking_news_indicator = 1").
		Order("urgency_score DESC, created_at DESC").
		Limit(limit).
		Find(&topics).Error
	return topics, err
}

// Analytics methods
func (r *analysisRepository) GetSentimentHourlyData(ctx context.Context, symbol string, hours int) ([]*model.SentimentHourlyMV, error) {
	var data []*model.SentimentHourlyMV
	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)

	err := r.db.WithContext(ctx).
		Where("primary_symbol = ? AND hour_timestamp >= ?", symbol, cutoff).
		Order("hour_timestamp ASC").
		Find(&data).Error
	return data, err
}

func (r *analysisRepository) GetTopTopicsLastDay(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	cutoff := time.Now().UTC().Add(-24 * time.Hour)

	var results []map[string]interface{}
	err := r.db.WithContext(ctx).
		Model(&model.TopicClassification{}).
		Select("primary_topic, COUNT(*) as article_count, AVG(primary_topic_confidence) as avg_confidence").
		Where("created_at >= ?", cutoff).
		Group("primary_topic").
		Order("article_count DESC").
		Limit(limit).
		Find(&results).Error

	return results, err
}

func (r *analysisRepository) GetMostMentionedEntities(ctx context.Context, days int, limit int) ([]map[string]interface{}, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)

	var results []map[string]interface{}
	err := r.db.WithContext(ctx).
		Model(&model.EntityRecognition{}).
		Select("entity_text, entity_type, COUNT(*) as mention_count, AVG(confidence) as avg_confidence").
		Where("created_at >= ? AND confidence > 0.7", cutoff).
		Group("entity_text, entity_type").
		Order("mention_count DESC").
		Limit(limit).
		Find(&results).Error

	return results, err
}
