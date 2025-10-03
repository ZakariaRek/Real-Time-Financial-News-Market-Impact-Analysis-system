// services/nlp-processing/internal/service/nlp_service.go
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/repository"
	"github.com/sirupsen/logrus"
)

type NLPProcessingService interface {
	ProcessArticle(ctx context.Context, article *model.Article) (*model.AnalysisResult, error)
	ProcessBatch(ctx context.Context, articles []*model.Article, options *model.ProcessingOptions) ([]*model.AnalysisResult, error)
	GetAnalysisResult(ctx context.Context, articleID string) (*model.AnalysisResult, error)
	InitializeModels(ctx context.Context) error
	GetModelStatus() *model.ModelStatus
}

type nlpProcessingService struct {
	sentimentService SentimentService
	analysisRepo     repository.AnalysisRepository
	cacheRepo        repository.CacheRepository
	modelStatus      *model.ModelStatus
}

func NewNLPProcessingService(
	sentimentService SentimentService,
	analysisRepo repository.AnalysisRepository,
	cacheRepo repository.CacheRepository,
) NLPProcessingService {
	return &nlpProcessingService{
		sentimentService: sentimentService,
		analysisRepo:     analysisRepo,
		cacheRepo:        cacheRepo,
		modelStatus: &model.ModelStatus{
			FinbertLoaded:  false,
			FinbertVersion: "textblob-1.0",
		},
	}
}

func (s *nlpProcessingService) InitializeModels(ctx context.Context) error {
	logrus.Info("Initializing NLP models for S&P 500 sentiment analysis...")

	if err := s.sentimentService.LoadModel(ctx); err != nil {
		return fmt.Errorf("failed to load sentiment model: %w", err)
	}

	s.modelStatus.FinbertLoaded = true
	logrus.Info("S&P 500 sentiment analysis models loaded successfully")
	return nil
}

func (s *nlpProcessingService) GetModelStatus() *model.ModelStatus {
	return s.modelStatus
}

func (s *nlpProcessingService) ProcessArticle(ctx context.Context, article *model.Article) (*model.AnalysisResult, error) {
	startTime := time.Now()

	// Check cache
	cachedResult, err := s.cacheRepo.GetAnalysisResult(ctx, article.ID.String())
	if err == nil && cachedResult != nil {
		return cachedResult, nil
	}

	result := &model.AnalysisResult{
		ArticleID: article.ID.String(),
		Status:    "processing",
		CreatedAt: time.Now().UTC(),
	}

	// Analyze sentiment
	fullText := article.Title + " " + article.Content
	sentiment, err := s.sentimentService.AnalyzeSentiment(ctx, article.ID.String(), fullText, article.Symbols)

	if err != nil {
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Sentiment analysis failed: %v", err)
	} else {
		result.SentimentAnalysis = sentiment
		result.Status = "completed"
	}

	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	// Store in database
	if err := s.analysisRepo.CreateSentimentAnalysis(ctx, sentiment); err != nil {
		logrus.WithError(err).Error("Failed to store sentiment analysis")
	}

	// Cache result
	if err := s.cacheRepo.SetAnalysisResult(ctx, article.ID.String(), result, time.Hour); err != nil {
		logrus.WithError(err).Warn("Failed to cache analysis result")
	}

	logrus.WithFields(logrus.Fields{
		"article_id":      article.ID,
		"processing_time": result.ProcessingTimeMs,
		"sentiment":       sentiment.CompoundScore,
	}).Info("Article sentiment analysis completed")

	return result, nil
}

func (s *nlpProcessingService) ProcessBatch(ctx context.Context, articles []*model.Article, options *model.ProcessingOptions) ([]*model.AnalysisResult, error) {
	results := make([]*model.AnalysisResult, len(articles))

	for i, article := range articles {
		result, err := s.ProcessArticle(ctx, article)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to process article %s", article.ID)
			result = &model.AnalysisResult{
				ArticleID:    article.ID.String(),
				Status:       "failed",
				ErrorMessage: err.Error(),
				CreatedAt:    time.Now().UTC(),
			}
		}
		results[i] = result
	}

	return results, nil
}

func (s *nlpProcessingService) GetAnalysisResult(ctx context.Context, articleID string) (*model.AnalysisResult, error) {
	// Check cache first
	result, err := s.cacheRepo.GetAnalysisResult(ctx, articleID)
	if err == nil && result != nil {
		return result, nil
	}

	// Reconstruct from database
	sentiment, err := s.analysisRepo.GetSentimentByArticleID(ctx, articleID)
	if err != nil {
		return nil, fmt.Errorf("analysis result not found for article %s: %w", articleID, err)
	}

	result = &model.AnalysisResult{
		ArticleID:         articleID,
		SentimentAnalysis: sentiment,
		Status:            "completed",
		CreatedAt:         sentiment.CreatedAt,
	}

	// Cache it
	s.cacheRepo.SetAnalysisResult(ctx, articleID, result, time.Hour)

	return result, nil
}
