// services/nlp-processing/internal/service/nlp_service.go
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/repository"
)

type NLPProcessingService interface {
	ProcessArticle(ctx context.Context, article *model.Article) (*model.AnalysisResult, error)
	ProcessBatch(ctx context.Context, articles []*model.Article, options *model.ProcessingOptions) ([]*model.AnalysisResult, error)
	GetAnalysisResult(ctx context.Context, articleID string) (*model.AnalysisResult, error)
	GetSentimentTrends(ctx context.Context, query *model.SentimentQuery) ([]*model.SentimentTrend, error)
	StreamProcessing(ctx context.Context, articleChan <-chan *model.Article, resultChan chan<- *model.AnalysisResult) error
	InitializeModels(ctx context.Context) error
	GetModelStatus() *model.ModelStatus
}

type nlpProcessingService struct {
	sentimentService SentimentService
	nerService       NERService
	topicService     TopicService
	analysisRepo     repository.AnalysisRepository
	cacheRepo        repository.CacheRepository

	workerCount       int
	batchSize         int
	processingTimeout time.Duration

	modelStatus *model.ModelStatus
	mu          sync.RWMutex
}

type NLPConfig struct {
	WorkerCount    int `mapstructure:"worker_count"`
	BatchSize      int `mapstructure:"batch_size"`
	TimeoutSeconds int `mapstructure:"timeout_seconds"`
	RetryAttempts  int `mapstructure:"retry_attempts"`
}

func NewNLPProcessingService(
	sentimentService SentimentService,
	nerService NERService,
	topicService TopicService,
	analysisRepo repository.AnalysisRepository,
	cacheRepo repository.CacheRepository,
	config NLPConfig,
) NLPProcessingService {
	return &nlpProcessingService{
		sentimentService:  sentimentService,
		nerService:        nerService,
		topicService:      topicService,
		analysisRepo:      analysisRepo,
		cacheRepo:         cacheRepo,
		workerCount:       config.WorkerCount,
		batchSize:         config.BatchSize,
		processingTimeout: time.Duration(config.TimeoutSeconds) * time.Second,
		modelStatus: &model.ModelStatus{
			FinbertLoaded:    false,
			NerModelLoaded:   false,
			TopicModelLoaded: false,
		},
	}
}

func (s *nlpProcessingService) InitializeModels(ctx context.Context) error {
	logrus.Info("Initializing NLP models...")

	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// Load models concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.sentimentService.LoadModel(ctx); err != nil {
			errChan <- fmt.Errorf("failed to load sentiment model: %w", err)
			return
		}
		s.mu.Lock()
		s.modelStatus.FinbertLoaded = true
		s.modelStatus.FinbertVersion = "finbert-1.0"
		s.mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.nerService.LoadModel(ctx); err != nil {
			errChan <- fmt.Errorf("failed to load NER model: %w", err)
			return
		}
		s.mu.Lock()
		s.modelStatus.NerModelLoaded = true
		s.modelStatus.NerVersion = "spacy-3.4"
		s.mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.topicService.LoadModel(ctx); err != nil {
			errChan <- fmt.Errorf("failed to load topic model: %w", err)
			return
		}
		s.mu.Lock()
		s.modelStatus.TopicModelLoaded = true
		s.modelStatus.TopicVersion = "cnn-lstm-1.0"
		s.mu.Unlock()
	}()

	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("model loading errors: %v", errors)
	}

	logrus.Info("All NLP models loaded successfully")
	return nil
}

func (s *nlpProcessingService) GetModelStatus() *model.ModelStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy to avoid race conditions
	return &model.ModelStatus{
		FinbertLoaded:    s.modelStatus.FinbertLoaded,
		NerModelLoaded:   s.modelStatus.NerModelLoaded,
		TopicModelLoaded: s.modelStatus.TopicModelLoaded,
		FinbertVersion:   s.modelStatus.FinbertVersion,
		NerVersion:       s.modelStatus.NerVersion,
		TopicVersion:     s.modelStatus.TopicVersion,
	}
}

func (s *nlpProcessingService) ProcessArticle(ctx context.Context, article *model.Article) (*model.AnalysisResult, error) {
	startTime := time.Now()

	// Check if result is already cached
	cachedResult, err := s.cacheRepo.GetAnalysisResult(ctx, article.ID.String())
	if err == nil && cachedResult != nil {
		logrus.WithField("article_id", article.ID).Debug("Returning cached analysis result")
		return cachedResult, nil
	}

	// Set processing status
	if err := s.cacheRepo.SetProcessingStatus(ctx, article.ID.String(), "processing", 10*time.Minute); err != nil {
		logrus.WithError(err).Warn("Failed to set processing status in cache")
	}

	result := &model.AnalysisResult{
		ArticleID: article.ID.String(),
		Status:    "processing",
	}

	// Create context with timeout
	processCtx, cancel := context.WithTimeout(ctx, s.processingTimeout)
	defer cancel()

	// Prepare text for analysis
	fullText := article.Title + " " + article.Content
	symbols := article.Symbols

	// Process all analyses concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Sentiment Analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		sentiment, err := s.sentimentService.AnalyzeSentiment(processCtx, article.ID.String(), fullText, symbols)

		mu.Lock()
		if err != nil {
			result.ErrorMessage += fmt.Sprintf("Sentiment analysis failed: %v; ", err)
		} else {
			result.SentimentAnalysis = sentiment
		}
		mu.Unlock()
	}()

	// Named Entity Recognition
	wg.Add(1)
	go func() {
		defer wg.Done()
		entities, err := s.nerService.ExtractEntities(processCtx, article.ID.String(), fullText)

		mu.Lock()
		if err != nil {
			result.ErrorMessage += fmt.Sprintf("NER failed: %v; ", err)
		} else {
			result.EntityRecognition = entities
		}
		mu.Unlock()
	}()

	// Topic Classification
	wg.Add(1)
	go func() {
		defer wg.Done()
		topic, err := s.topicService.ClassifyTopic(processCtx, article.ID.String(), fullText)

		mu.Lock()
		if err != nil {
			result.ErrorMessage += fmt.Sprintf("Topic classification failed: %v; ", err)
		} else {
			result.TopicClassification = topic
		}
		mu.Unlock()
	}()

	wg.Wait()

	// Calculate processing time
	processingTime := time.Since(startTime)
	result.ProcessingTimeMs = processingTime.Nanoseconds() / int64(time.Millisecond)
	result.CreatedAt = time.Now().UTC()

	// Determine final status
	if result.ErrorMessage == "" {
		result.Status = "completed"
	} else {
		result.Status = "partial_failure"
		logrus.WithFields(logrus.Fields{
			"article_id": article.ID,
			"errors":     result.ErrorMessage,
		}).Warn("Article processing completed with errors")
	}

	// Store results in database
	if err := s.storeAnalysisResults(ctx, result); err != nil {
		logrus.WithError(err).Error("Failed to store analysis results in database")
	}

	// Cache the result
	if err := s.cacheRepo.SetAnalysisResult(ctx, article.ID.String(), result, time.Hour); err != nil {
		logrus.WithError(err).Warn("Failed to cache analysis result")
	}

	logrus.WithFields(logrus.Fields{
		"article_id":      article.ID,
		"processing_time": processingTime.Milliseconds(),
		"status":          result.Status,
	}).Info("Article processing completed")

	return result, nil
}

func (s *nlpProcessingService) ProcessBatch(ctx context.Context, articles []*model.Article, options *model.ProcessingOptions) ([]*model.AnalysisResult, error) {
	results := make([]*model.AnalysisResult, 0, len(articles))

	// Process articles in batches to manage memory and resources
	for i := 0; i < len(articles); i += s.batchSize {
		end := i + s.batchSize
		if end > len(articles) {
			end = len(articles)
		}

		batch := articles[i:end]
		batchResults, err := s.processBatch(ctx, batch, options)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to process batch %d-%d", i, end)
			continue
		}

		results = append(results, batchResults...)
	}

	return results, nil
}

func (s *nlpProcessingService) processBatch(ctx context.Context, batch []*model.Article, options *model.ProcessingOptions) ([]*model.AnalysisResult, error) {
	results := make([]*model.AnalysisResult, len(batch))

	// Process articles concurrently within the batch
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.workerCount)

	for i, article := range batch {
		wg.Add(1)
		go func(idx int, art *model.Article) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result, err := s.ProcessArticle(ctx, art)
			if err != nil {
				logrus.WithError(err).Errorf("Failed to process article %s", art.ID)
				result = &model.AnalysisResult{
					ArticleID:    art.ID.String(),
					Status:       "failed",
					ErrorMessage: err.Error(),
					CreatedAt:    time.Now().UTC(),
				}
			}

			results[idx] = result
		}(i, article)
	}

	wg.Wait()
	return results, nil
}

func (s *nlpProcessingService) GetAnalysisResult(ctx context.Context, articleID string) (*model.AnalysisResult, error) {
	// First check cache
	result, err := s.cacheRepo.GetAnalysisResult(ctx, articleID)
	if err == nil && result != nil {
		return result, nil
	}

	// If not in cache, try to reconstruct from database
	result, err = s.reconstructAnalysisResult(ctx, articleID)
	if err != nil {
		return nil, fmt.Errorf("analysis result not found for article %s: %w", articleID, err)
	}

	// Cache the reconstructed result
	if err := s.cacheRepo.SetAnalysisResult(ctx, articleID, result, time.Hour); err != nil {
		logrus.WithError(err).Warn("Failed to cache reconstructed analysis result")
	}

	return result, nil
}

func (s *nlpProcessingService) GetSentimentTrends(ctx context.Context, query *model.SentimentQuery) ([]*model.SentimentTrend, error) {
	// Delegate to sentiment service for now
	// In a full implementation, this might involve more complex analytics
	return s.sentimentService.GetSentimentTrends(ctx, query.Symbol, query.StartTime, query.EndTime)
}

func (s *nlpProcessingService) StreamProcessing(ctx context.Context, articleChan <-chan *model.Article, resultChan chan<- *model.AnalysisResult) error {
	// Create worker pool for stream processing
	var wg sync.WaitGroup

	for i := 0; i < s.workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case article, ok := <-articleChan:
					if !ok {
						return // Channel closed
					}

					result, err := s.ProcessArticle(ctx, article)
					if err != nil {
						logrus.WithError(err).Errorf("Stream processing failed for article %s", article.ID)
						result = &model.AnalysisResult{
							ArticleID:    article.ID.String(),
							Status:       "failed",
							ErrorMessage: err.Error(),
							CreatedAt:    time.Now().UTC(),
						}
					}

					select {
					case resultChan <- result:
					case <-ctx.Done():
						return
					}

				case <-ctx.Done():
					return
				}
			}
		}()
	}

	wg.Wait()
	return nil
}

func (s *nlpProcessingService) storeAnalysisResults(ctx context.Context, result *model.AnalysisResult) error {
	// Store sentiment analysis
	if result.SentimentAnalysis != nil {
		if err := s.analysisRepo.CreateSentimentAnalysis(ctx, result.SentimentAnalysis); err != nil {
			return fmt.Errorf("failed to store sentiment analysis: %w", err)
		}
	}

	// Store entity recognition results
	if len(result.EntityRecognition) > 0 {
		if err := s.analysisRepo.CreateEntityRecognition(ctx, result.EntityRecognition); err != nil {
			return fmt.Errorf("failed to store entity recognition: %w", err)
		}
	}

	// Store topic classification
	if result.TopicClassification != nil {
		if err := s.analysisRepo.CreateTopicClassification(ctx, result.TopicClassification); err != nil {
			return fmt.Errorf("failed to store topic classification: %w", err)
		}
	}

	return nil
}

func (s *nlpProcessingService) reconstructAnalysisResult(ctx context.Context, articleID string) (*model.AnalysisResult, error) {
	result := &model.AnalysisResult{
		ArticleID: articleID,
		Status:    "completed",
		CreatedAt: time.Now().UTC(),
	}

	// Get sentiment analysis
	sentiment, err := s.analysisRepo.GetSentimentByArticleID(ctx, articleID)
	if err == nil {
		result.SentimentAnalysis = sentiment
	}

	// Get entity recognition
	//entities, err := s.analysisRepo.GetEntitiesByArticleID(ctx, articleID)
	//if err == nil {
	//	result.EntityRecognition = entities
	//}

	// Get topic classification
	topic, err := s.analysisRepo.GetTopicByArticleID(ctx, articleID)
	if err == nil {
		result.TopicClassification = topic
	}

	// If nothing was found, return error
	if result.SentimentAnalysis == nil && len(result.EntityRecognition) == 0 && result.TopicClassification == nil {
		return nil, fmt.Errorf("no analysis data found for article %s", articleID)
	}

	return result, nil
}
