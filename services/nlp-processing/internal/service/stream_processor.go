// services/nlp-processing/internal/service/stream_processor.go
package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	newsv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/proto/services/news-ingestion/proto/gen"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/client"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/repository"
)

type StreamProcessor struct {
	newsClient        *client.NewsIngestionClient
	nlpService        NLPProcessingService
	articleRepo       repository.ArticleRepository
	processingLogRepo repository.ProcessingLogRepository

	workerCount    int
	batchSize      int32
	processedCount int64
	failedCount    int64
	mu             sync.RWMutex
}

type StreamProcessorConfig struct {
	NewsIngestionEndpoint string
	WorkerCount           int
	BatchSize             int32
}

func NewStreamProcessor(
	config StreamProcessorConfig,
	nlpService NLPProcessingService,
	articleRepo repository.ArticleRepository,
	processingLogRepo repository.ProcessingLogRepository,
) (*StreamProcessor, error) {
	newsClient, err := client.NewNewsIngestionClient(config.NewsIngestionEndpoint)
	if err != nil {
		return nil, err
	}

	return &StreamProcessor{
		newsClient:        newsClient,
		nlpService:        nlpService,
		articleRepo:       articleRepo,
		processingLogRepo: processingLogRepo,
		workerCount:       config.WorkerCount,
		batchSize:         config.BatchSize,
	}, nil
}

func (sp *StreamProcessor) Start(ctx context.Context) error {
	logrus.Info("Starting NLP stream processor...")

	// Create worker pool
	articleChan, errorChan := sp.newsClient.StreamPendingArticles(ctx, sp.batchSize)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < sp.workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			sp.worker(ctx, workerID, articleChan)
		}(i)
	}

	// Monitor errors
	go func() {
		for err := range errorChan {
			logrus.WithError(err).Error("Stream error occurred")
		}
	}()

	// Log statistics periodically
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				sp.logStatistics()
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
	return nil
}

func (sp *StreamProcessor) worker(ctx context.Context, workerID int, articleChan <-chan *newsv1.Article) {
	logrus.WithField("worker_id", workerID).Info("Worker started")

	for {
		select {
		case article, ok := <-articleChan:
			if !ok {
				logrus.WithField("worker_id", workerID).Info("Article channel closed, worker stopping")
				return
			}

			if err := sp.processArticle(ctx, article); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"worker_id":  workerID,
					"article_id": article.Id,
				}).Error("Failed to process article")
				sp.incrementFailed()
			} else {
				sp.incrementProcessed()
			}

		case <-ctx.Done():
			logrus.WithField("worker_id", workerID).Info("Context cancelled, worker stopping")
			return
		}
	}
}

func (sp *StreamProcessor) processArticle(ctx context.Context, protoArticle *newsv1.Article) error {
	startTime := time.Now()

	// Convert proto to model
	article := sp.protoToModel(protoArticle)

	// Process with NLP service
	result, err := sp.nlpService.ProcessArticle(ctx, article)

	processingTime := time.Since(startTime).Milliseconds()
	success := err == nil && result.Status != "failed"
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	} else if result.ErrorMessage != "" {
		errorMsg = result.ErrorMessage
	}

	// Acknowledge to news-ingestion service
	if ackErr := sp.newsClient.AcknowledgeProcessing(
		ctx,
		article.ID.String(),
		success,
		errorMsg,
		processingTime,
	); ackErr != nil {
		logrus.WithError(ackErr).Warn("Failed to acknowledge article processing")
	}

	// Log processing
	sp.logArticleProcessing(ctx, article.ID, success, processingTime, errorMsg)

	return err
}

func (sp *StreamProcessor) protoToModel(proto *newsv1.Article) *model.Article {
	id, _ := uuid.Parse(proto.Id)
	if id == uuid.Nil {
		id = uuid.New()
	}

	return &model.Article{
		ID:          id,
		SourceID:    uint(proto.SourceId),
		Title:       proto.Title,
		Content:     proto.Content,
		URL:         proto.Url,
		Symbols:     proto.Symbols,
		PublishedAt: proto.PublishedAt.AsTime(),
	}
}

func (sp *StreamProcessor) logArticleProcessing(ctx context.Context, articleID uuid.UUID, success bool, processingTime int64, errorMsg string) {
	status := "completed"
	if !success {
		status = "failed"
	}

	log := &model.ArticleProcessingLog{
		ArticleID:        articleID,
		ProcessingStage:  "nlp_processing",
		Status:           status,
		ProcessingTimeMs: int(processingTime),
		ErrorMessage:     errorMsg,
	}

	if err := sp.processingLogRepo.Create(ctx, log); err != nil {
		logrus.WithError(err).Error("Failed to create processing log")
	}
}

func (sp *StreamProcessor) incrementProcessed() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.processedCount++
}

func (sp *StreamProcessor) incrementFailed() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.failedCount++
}

func (sp *StreamProcessor) logStatistics() {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	logrus.WithFields(logrus.Fields{
		"processed": sp.processedCount,
		"failed":    sp.failedCount,
		"success_rate": func() float64 {
			total := sp.processedCount + sp.failedCount
			if total == 0 {
				return 0
			}
			return float64(sp.processedCount) / float64(total) * 100
		}(),
	}).Info("Stream processor statistics")
}

func (sp *StreamProcessor) Close() error {
	if sp.newsClient != nil {
		return sp.newsClient.Close()
	}
	return nil
}
