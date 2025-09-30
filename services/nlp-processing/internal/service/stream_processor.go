// services/nlp-processing/internal/service/stream_processor.go
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	newsv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/proto/services/news-ingestion/proto/gen"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/client"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
)

type StreamProcessor struct {
	newsClient     *client.NewsIngestionClient
	nlpService     NLPProcessingService
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
) (*StreamProcessor, error) {
	newsClient, err := client.NewNewsIngestionClient(config.NewsIngestionEndpoint)
	if err != nil {
		return nil, err
	}

	return &StreamProcessor{
		newsClient:  newsClient,
		nlpService:  nlpService,
		workerCount: config.WorkerCount,
		batchSize:   config.BatchSize,
	}, nil
}

func (sp *StreamProcessor) Start(ctx context.Context) error {
	logrus.Info("Starting S&P 500 sentiment analysis stream processor...")

	articleChan, errorChan := sp.newsClient.StreamPendingArticles(ctx, sp.batchSize)

	var wg sync.WaitGroup
	for i := 0; i < sp.workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			sp.worker(ctx, workerID, articleChan)
		}(i)
	}

	go func() {
		for err := range errorChan {
			logrus.WithError(err).Error("Stream error occurred")
		}
	}()

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
	logrus.WithField("worker_id", workerID).Info("Sentiment analysis worker started")

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

	article := sp.protoToModel(protoArticle)
	result, err := sp.nlpService.ProcessArticle(ctx, article)

	processingTime := time.Since(startTime).Milliseconds()
	success := err == nil && result.Status != "failed"
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	} else if result.ErrorMessage != "" {
		errorMsg = result.ErrorMessage
	}

	if ackErr := sp.newsClient.AcknowledgeProcessing(
		ctx,
		article.ID.String(),
		success,
		errorMsg,
		processingTime,
	); ackErr != nil {
		logrus.WithError(ackErr).Warn("Failed to acknowledge article processing")
	}

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

	total := sp.processedCount + sp.failedCount
	successRate := 0.0
	if total > 0 {
		successRate = float64(sp.processedCount) / float64(total) * 100
	}

	logrus.WithFields(logrus.Fields{
		"processed":    sp.processedCount,
		"failed":       sp.failedCount,
		"success_rate": fmt.Sprintf("%.2f%%", successRate),
	}).Info("S&P 500 sentiment analysis statistics")
}

func (sp *StreamProcessor) Close() error {
	if sp.newsClient != nil {
		return sp.newsClient.Close()
	}
	return nil
}
