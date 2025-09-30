// services/news-ingestion/internal/handler/grpc_streaming_handler.go
package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
	_ "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/repository"
	newsv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/proto/services/news-ingestion/proto/gen"
)

// StreamArticles streams pending articles to NLP Processing Service
func (h *GRPCHandler) StreamArticles(req *newsv1.StreamArticlesRequest, stream newsv1.NewsService_StreamArticlesServer) error {
	ctx := stream.Context()
	logrus.Infof("gRPC StreamArticles called - BatchSize: %d, Status: %s", req.BatchSize, req.Status)

	// Set defaults
	batchSize := int(req.BatchSize)
	if batchSize <= 0 || batchSize > 100 {
		batchSize = 10
	}

	status := req.Status
	if status == "" {
		status = "pending"
	}

	// Stream articles continuously or as a single batch
	if req.Continuous {
		return h.streamArticlesContinuously(ctx, stream, batchSize, status, req.Symbols)
	}

	return h.streamArticlesBatch(ctx, stream, batchSize, status, req.Symbols)
}

// streamArticlesBatch sends a single batch of articles
func (h *GRPCHandler) streamArticlesBatch(
	ctx context.Context,
	stream newsv1.NewsService_StreamArticlesServer,
	batchSize int,
	//status string,
	articleStatus string, // renamed from 'status'

	symbols []string,
) error {
	var articles []*model.Article
	var err error

	// Fetch articles based on criteria
	if len(symbols) > 0 {
		articles, err = h.articleRepo.GetBySymbols(ctx, symbols, batchSize)
	} else if articleStatus == "pending" {
		articles, err = h.articleRepo.GetPendingArticles(ctx, batchSize)
	} else {
		// Get recent articles
		end := time.Now().UTC()
		start := end.Add(-24 * time.Hour)
		articles, err = h.articleRepo.GetByDateRange(ctx, start, end)
		if len(articles) > batchSize {
			articles = articles[:batchSize]
		}
	}

	if err != nil {
		//logrus.WithError(err).Error("Failed to fetch articles for streaming")
		//return status.Error(codes.Internal, "failed to fetch articles")
		logrus.WithError(err).Error("Failed to fetch articles for streaming")
		return status.Error(codes.Internal, "failed to fetch articles")
	}

	logrus.Infof("Streaming %d articles", len(articles))

	// Stream each article
	for i, article := range articles {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			logrus.Info("Stream cancelled by client")
			return status.Error(codes.Canceled, "stream cancelled")
		default:
		}

		// Convert to proto and send
		protoArticle := h.modelToProtoArticle(article)

		response := &newsv1.StreamArticlesResponse{
			Article:        protoArticle,
			SequenceNumber: int32(i + 1),
			TotalCount:     int32(len(articles)),
			Timestamp:      timestamppb.Now(),
		}

		if err := stream.Send(response); err != nil {
			logrus.WithError(err).Errorf("Failed to stream article: %s", article.ID)
			return status.Error(codes.Internal, "failed to send article")
		}

		// Mark article as processing
		h.articleRepo.UpdateStatus(ctx, article.ID, model.StatusProcessing)

		logrus.Debugf("Streamed article %d/%d: %s", i+1, len(articles), article.Title)
	}

	logrus.Infof("Successfully streamed %d articles", len(articles))
	return nil
}

// streamArticlesContinuously streams articles continuously as they become available
func (h *GRPCHandler) streamArticlesContinuously(
	ctx context.Context,
	stream newsv1.NewsService_StreamArticlesServer,
	batchSize int,
	Astatus string,
	symbols []string,
) error {
	logrus.Info("Starting continuous article streaming")

	ticker := time.NewTicker(5 * time.Second) // Poll every 5 seconds
	defer ticker.Stop()

	sequenceNumber := int32(0)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Continuous stream cancelled by client")
			return status.Error(codes.Canceled, "stream cancelled")

		case <-ticker.C:
			// Fetch new articles
			var articles []*model.Article
			var err error

			if len(symbols) > 0 {
				articles, err = h.articleRepo.GetBySymbols(ctx, symbols, batchSize)
			} else {
				articles, err = h.articleRepo.GetPendingArticles(ctx, batchSize)
			}

			if err != nil {
				logrus.WithError(err).Warn("Failed to fetch articles in continuous stream")
				continue
			}

			if len(articles) == 0 {
				logrus.Debug("No new articles to stream")
				continue
			}

			logrus.Infof("Streaming batch of %d articles", len(articles))

			// Stream articles
			for _, article := range articles {
				sequenceNumber++

				protoArticle := h.modelToProtoArticle(article)

				response := &newsv1.StreamArticlesResponse{
					Article:        protoArticle,
					SequenceNumber: sequenceNumber,
					Timestamp:      timestamppb.Now(),
				}

				if err := stream.Send(response); err != nil {
					logrus.WithError(err).Error("Failed to send article in continuous stream")
					return status.Error(codes.Internal, "failed to send article")
				}

				// Mark as processing
				h.articleRepo.UpdateStatus(ctx, article.ID, model.StatusProcessing)

				logrus.Debugf("Streamed article #%d: %s", sequenceNumber, article.Title)
			}
		}
	}
}

// StreamArticlesByDateRange streams articles within a specific date range
func (h *GRPCHandler) StreamArticlesByDateRange(req *newsv1.StreamArticlesByDateRangeRequest, stream newsv1.NewsService_StreamArticlesByDateRangeServer) error {
	ctx := stream.Context()
	logrus.Infof("gRPC StreamArticlesByDateRange called - From: %v, To: %v", req.StartDate, req.EndDate)

	if req.StartDate == nil || req.EndDate == nil {
		return status.Error(codes.InvalidArgument, "start_date and end_date are required")
	}

	startDate := req.StartDate.AsTime()
	endDate := req.EndDate.AsTime()

	// Fetch articles in date range
	articles, err := h.articleRepo.GetByDateRange(ctx, startDate, endDate)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch articles by date range")
		return status.Error(codes.Internal, "failed to fetch articles")
	}

	logrus.Infof("Streaming %d articles from date range", len(articles))

	// Stream articles in batches
	batchSize := int(req.BatchSize)
	if batchSize <= 0 {
		batchSize = 50
	}

	for i := 0; i < len(articles); i += batchSize {
		// Check context
		select {
		case <-ctx.Done():
			return status.Error(codes.Canceled, "stream cancelled")
		default:
		}

		end := i + batchSize
		if end > len(articles) {
			end = len(articles)
		}

		batch := articles[i:end]

		for j, article := range batch {
			protoArticle := h.modelToProtoArticle(article)

			response := &newsv1.StreamArticlesByDateRangeResponse{
				Article:        protoArticle,
				SequenceNumber: int32(i + j + 1),
				TotalCount:     int32(len(articles)),
				Timestamp:      timestamppb.Now(),
			}

			if err := stream.Send(response); err != nil {
				logrus.WithError(err).Error("Failed to stream article")
				return status.Error(codes.Internal, "failed to send article")
			}
		}

		logrus.Debugf("Streamed batch %d/%d", end, len(articles))
	}

	logrus.Infof("Successfully streamed %d articles from date range", len(articles))
	return nil
}

// AcknowledgeArticleProcessing allows NLP service to acknowledge article processing
func (h *GRPCHandler) AcknowledgeArticleProcessing(ctx context.Context, req *newsv1.AcknowledgeArticleProcessingRequest) (*newsv1.AcknowledgeArticleProcessingResponse, error) {
	logrus.Infof("gRPC AcknowledgeArticleProcessing called for article: %s", req.ArticleId)

	articleID, err := uuid.Parse(req.ArticleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid article ID")
	}

	// Update article status based on acknowledgment
	var newStatus model.ProcessingStatus
	if req.Success {
		newStatus = model.StatusCompleted
	} else {
		newStatus = model.StatusFailed
	}

	if err := h.articleRepo.UpdateStatus(ctx, articleID, newStatus); err != nil {
		logrus.WithError(err).Error("Failed to update article status")
		return nil, status.Error(codes.Internal, "failed to update status")
	}

	// Log processing result
	if req.ErrorMessage != "" {
		h.logRepo.Create(ctx, &model.ArticleProcessingLog{
			ArticleID:        articleID,
			ProcessingStage:  "nlp_processing",
			Status:           "failed",
			ErrorMessage:     req.ErrorMessage,
			ProcessingTimeMs: int(req.ProcessingTimeMs),
		})
	} else {
		h.logRepo.Create(ctx, &model.ArticleProcessingLog{
			ArticleID:        articleID,
			ProcessingStage:  "nlp_processing",
			Status:           "completed",
			ProcessingTimeMs: int(req.ProcessingTimeMs),
		})
	}

	logrus.Infof("Article %s processing acknowledged: success=%v", req.ArticleId, req.Success)

	return &newsv1.AcknowledgeArticleProcessingResponse{
		Success: true,
		Message: "Processing acknowledged successfully",
	}, nil
}
