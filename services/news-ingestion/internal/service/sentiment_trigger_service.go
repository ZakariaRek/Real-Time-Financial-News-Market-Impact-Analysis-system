// services/news-ingestion/internal/service/sentiment_trigger_service.go
package service

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/client"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/repository"
)

const (
	MaxBatchSize = 100 // Maximum batch size allowed by NLP service
)

type SentimentTriggerService struct {
	articleRepo   repository.ArticleRepository
	nlpClient     client.NLPProcessingClient
	threshold     int
	checkInterval time.Duration
}

func NewSentimentTriggerService(
	articleRepo repository.ArticleRepository,
	nlpClient client.NLPProcessingClient,
	threshold int,
) *SentimentTriggerService {
	return &SentimentTriggerService{
		articleRepo:   articleRepo,
		nlpClient:     nlpClient,
		threshold:     threshold,
		checkInterval: 1 * time.Minute,
	}
}

// Start begins monitoring pending articles
func (s *SentimentTriggerService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	logrus.Info("Sentiment trigger service started")

	for {
		select {
		case <-ticker.C:
			if err := s.checkAndTrigger(ctx); err != nil {
				logrus.WithError(err).Error("Failed to check and trigger sentiment analysis")
			}
		case <-ctx.Done():
			logrus.Info("Sentiment trigger service stopped")
			return
		}
	}
}

// checkAndTrigger checks pending count and triggers if threshold met
func (s *SentimentTriggerService) checkAndTrigger(ctx context.Context) error {
	// Get count of pending articles
	pendingArticles, err := s.articleRepo.GetPendingArticles(ctx, 1000)
	if err != nil {
		return err
	}

	pendingCount := len(pendingArticles)

	logrus.WithField("pending_count", pendingCount).Debug("Checking pending articles")

	// Trigger if threshold exceeded
	if pendingCount >= s.threshold {
		logrus.WithFields(logrus.Fields{
			"pending_count": pendingCount,
			"threshold":     s.threshold,
		}).Info("✅ Threshold exceeded, triggering batch sentiment analysis")

		return s.triggerBatchAnalysis(ctx, pendingArticles)
	}

	return nil
}

// triggerBatchAnalysis sends articles to NLP service in batches
func (s *SentimentTriggerService) triggerBatchAnalysis(ctx context.Context, articles []*model.Article) error {
	totalArticles := len(articles)

	if totalArticles == 0 {
		return nil
	}

	// Calculate number of batches needed
	numBatches := (totalArticles + MaxBatchSize - 1) / MaxBatchSize

	logrus.WithFields(logrus.Fields{
		"total_articles": totalArticles,
		"num_batches":    numBatches,
		"batch_size":     MaxBatchSize,
	}).Info("📦 Splitting articles into batches for processing")

	startTime := time.Now()
	successfulBatches := 0
	failedBatches := 0
	totalProcessed := 0

	// Process articles in batches of MaxBatchSize
	for i := 0; i < totalArticles; i += MaxBatchSize {
		end := i + MaxBatchSize
		if end > totalArticles {
			end = totalArticles
		}

		batch := articles[i:end]
		batchNum := (i / MaxBatchSize) + 1

		logrus.WithFields(logrus.Fields{
			"batch":         batchNum,
			"total_batches": numBatches,
			"batch_size":    len(batch),
		}).Info("🔄 Processing batch...")

		// Send batch to NLP service
		batchStartTime := time.Now()
		if err := s.nlpClient.ProcessBatch(ctx, batch); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"batch":      batchNum,
				"batch_size": len(batch),
			}).Error("❌ Failed to process batch")
			failedBatches++
			continue
		}

		batchDuration := time.Since(batchStartTime)
		successfulBatches++
		totalProcessed += len(batch)

		logrus.WithFields(logrus.Fields{
			"batch":         batchNum,
			"batch_size":    len(batch),
			"duration":      batchDuration,
			"articles_left": totalArticles - end,
		}).Info("✅ Batch processed successfully")

		// Small delay between batches to avoid overwhelming the NLP service
		if end < totalArticles {
			time.Sleep(500 * time.Millisecond)
		}
	}

	totalDuration := time.Since(startTime)

	// Log final summary
	logrus.WithFields(logrus.Fields{
		"total_articles":     totalArticles,
		"successful_batches": successfulBatches,
		"failed_batches":     failedBatches,
		"articles_processed": totalProcessed,
		"total_duration":     totalDuration,
		"avg_per_article":    totalDuration / time.Duration(totalProcessed),
	}).Info("📊 Batch sentiment analysis completed")

	if failedBatches > 0 {
		return nil // Don't return error as some batches succeeded
	}

	return nil
}
