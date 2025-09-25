// services/nlp-processing/internal/handler/nlp_grpc_handler.go
package handler

import (
	"context"
	_ "fmt"
	"io"
	_ "time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/repository"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/service"
	nlpv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/proto/gen/nlp/v1"
)

type NLPGRPCHandler struct {
	nlpv1.UnimplementedNLPProcessingServiceServer
	nlpService   service.NLPProcessingService
	analysisRepo repository.AnalysisRepository
}

func NewNLPGRPCHandler(
	nlpService service.NLPProcessingService,
	analysisRepo repository.AnalysisRepository,
) *NLPGRPCHandler {
	return &NLPGRPCHandler{
		nlpService:   nlpService,
		analysisRepo: analysisRepo,
	}
}

func (h *NLPGRPCHandler) StreamArticles(stream nlpv1.NLPProcessingService_StreamArticlesServer) error {
	logrus.Info("StreamArticles gRPC method called")

	articleChan := make(chan *model.Article, 100)
	resultChan := make(chan *model.AnalysisResult, 100)

	ctx := stream.Context()

	// Start processing goroutine
	go func() {
		defer close(resultChan)
		if err := h.nlpService.StreamProcessing(ctx, articleChan, resultChan); err != nil {
			logrus.WithError(err).Error("Stream processing failed")
		}
	}()

	// Start sending results goroutine
	go func() {
		for result := range resultChan {
			protoResult := h.modelToProtoAnalysisResult(result)
			if err := stream.Send(protoResult); err != nil {
				logrus.WithError(err).Error("Failed to send analysis result")
				return
			}
		}
	}()

	// Receive articles from stream
	for {
		article, err := stream.Recv()
		if err == io.EOF {
			close(articleChan)
			break
		}
		if err != nil {
			logrus.WithError(err).Error("Failed to receive article from stream")
			return status.Error(codes.Internal, "failed to receive article")
		}

		modelArticle := h.protoToModelArticle(article)
		select {
		case articleChan <- modelArticle:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (h *NLPGRPCHandler) ProcessArticle(ctx context.Context, req *nlpv1.ProcessArticleRequest) (*nlpv1.ProcessArticleResponse, error) {
	logrus.WithField("article_id", req.Article.Id).Info("ProcessArticle gRPC method called")

	if req.Article == nil {
		return nil, status.Error(codes.InvalidArgument, "article is required")
	}

	// Convert proto to model
	article := h.protoToModelArticle(req.Article)

	// Process the article
	result, err := h.nlpService.ProcessArticle(ctx, article)
	if err != nil {
		logrus.WithError(err).Error("Failed to process article")
		return nil, status.Error(codes.Internal, "failed to process article")
	}

	return &nlpv1.ProcessArticleResponse{
		Result:  h.modelToProtoAnalysisResult(result),
		Success: result.Status != "failed",
		Message: "Article processed successfully",
	}, nil
}

func (h *NLPGRPCHandler) ProcessBatch(ctx context.Context, req *nlpv1.BatchProcessRequest) (*nlpv1.BatchProcessResponse, error) {
	logrus.WithField("batch_size", len(req.Articles)).Info("ProcessBatch gRPC method called")

	if len(req.Articles) == 0 {
		return nil, status.Error(codes.InvalidArgument, "articles list cannot be empty")
	}

	if len(req.Articles) > 100 {
		return nil, status.Error(codes.InvalidArgument, "batch size cannot exceed 100")
	}

	// Convert proto to model
	articles := make([]*model.Article, len(req.Articles))
	for i, protoArticle := range req.Articles {
		articles[i] = h.protoToModelArticle(protoArticle)
	}

	// Process the batch
	results, err := h.nlpService.ProcessBatch(ctx, articles, h.protoToModelProcessingOptions(req.Options))
	if err != nil {
		logrus.WithError(err).Error("Failed to process batch")
		return nil, status.Error(codes.Internal, "failed to process batch")
	}

	// Convert results to proto
	protoResults := make([]*nlpv1.AnalysisResult, len(results))
	successfulCount := int32(0)
	failedCount := int32(0)
	var errors []string

	for i, result := range results {
		protoResults[i] = h.modelToProtoAnalysisResult(result)
		if result.Status == "failed" {
			failedCount++
			if result.ErrorMessage != "" {
				errors = append(errors, result.ErrorMessage)
			}
		} else {
			successfulCount++
		}
	}

	return &nlpv1.BatchProcessResponse{
		Results:         protoResults,
		SuccessfulCount: successfulCount,
		FailedCount:     failedCount,
		Errors:          errors,
	}, nil
}

func (h *NLPGRPCHandler) GetAnalysisResult(ctx context.Context, req *nlpv1.GetAnalysisRequest) (*nlpv1.GetAnalysisResponse, error) {
	logrus.WithField("article_id", req.ArticleId).Info("GetAnalysisResult gRPC method called")

	if req.ArticleId == "" {
		return nil, status.Error(codes.InvalidArgument, "article_id is required")
	}

	result, err := h.nlpService.GetAnalysisResult(ctx, req.ArticleId)
	if err != nil {
		logrus.WithError(err).Error("Failed to get analysis result")
		return &nlpv1.GetAnalysisResponse{
			Result: nil,
			Found:  false,
		}, nil
	}

	return &nlpv1.GetAnalysisResponse{
		Result: h.modelToProtoAnalysisResult(result),
		Found:  true,
	}, nil
}

func (h *NLPGRPCHandler) GetSentimentTrends(ctx context.Context, req *nlpv1.SentimentTrendsRequest) (*nlpv1.SentimentTrendsResponse, error) {
	logrus.WithField("symbol", req.Symbol).Info("GetSentimentTrends gRPC method called")

	if req.Symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	query := &model.SentimentQuery{
		Symbol:    req.Symbol,
		StartTime: req.StartTime.AsTime(),
		EndTime:   req.EndTime.AsTime(),
		Interval:  req.Interval,
	}

	trends, err := h.nlpService.GetSentimentTrends(ctx, query)
	if err != nil {
		logrus.WithError(err).Error("Failed to get sentiment trends")
		return nil, status.Error(codes.Internal, "failed to get sentiment trends")
	}

	// Convert to proto
	protoTrends := make([]*nlpv1.SentimentTrend, len(trends))
	for i, trend := range trends {
		protoTrends[i] = &nlpv1.SentimentTrend{
			Timestamp:    timestamppb.New(trend.Timestamp),
			Symbol:       trend.Symbol,
			AvgSentiment: trend.AvgSentiment,
			ArticleCount: trend.ArticleCount,
			Volatility:   trend.Volatility,
		}
	}

	return &nlpv1.SentimentTrendsResponse{
		Trends:     protoTrends,
		TotalCount: int32(len(trends)),
	}, nil
}

func (h *NLPGRPCHandler) GetEntityMentions(ctx context.Context, req *nlpv1.EntityMentionsRequest) (*nlpv1.EntityMentionsResponse, error) {
	logrus.WithField("entity", req.EntityText).Info("GetEntityMentions gRPC method called")

	// This would typically query the analysis repository for entity mentions
	// For now, return empty response
	return &nlpv1.EntityMentionsResponse{
		Mentions:   []*nlpv1.EntityMention{},
		TotalCount: 0,
	}, nil
}

func (h *NLPGRPCHandler) GetTopicDistribution(ctx context.Context, req *nlpv1.TopicDistributionRequest) (*nlpv1.TopicDistributionResponse, error) {
	logrus.Info("GetTopicDistribution gRPC method called")

	// This would typically query the analysis repository for topic distribution
	// For now, return stub data
	return &nlpv1.TopicDistributionResponse{
		Topics: []*nlpv1.TopicDistribution{},
	}, nil
}

func (h *NLPGRPCHandler) StreamSentimentUpdates(req *nlpv1.SentimentStreamRequest, stream nlpv1.NLPProcessingService_StreamSentimentUpdatesServer) error {
	logrus.Info("StreamSentimentUpdates gRPC method called")

	// This would implement real-time sentiment streaming
	// For now, return empty implementation
	return nil
}

func (h *NLPGRPCHandler) StreamBreakingNews(req *nlpv1.BreakingNewsRequest, stream nlpv1.NLPProcessingService_StreamBreakingNewsServer) error {
	logrus.Info("StreamBreakingNews gRPC method called")

	// This would implement real-time breaking news streaming
	// For now, return empty implementation
	return nil
}

func (h *NLPGRPCHandler) HealthCheck(ctx context.Context, req *emptypb.Empty) (*nlpv1.HealthCheckResponse, error) {
	logrus.Info("HealthCheck gRPC method called")

	modelStatus := h.nlpService.GetModelStatus()

	return &nlpv1.HealthCheckResponse{
		Status:         "healthy",
		Service:        "nlp-processing",
		Timestamp:      timestamppb.Now(),
		DatabaseStatus: "connected",
		ModelStatus: &nlpv1.ModelStatus{
			FinbertLoaded:    modelStatus.FinbertLoaded,
			NerModelLoaded:   modelStatus.NerModelLoaded,
			TopicModelLoaded: modelStatus.TopicModelLoaded,
			FinbertVersion:   modelStatus.FinbertVersion,
			NerVersion:       modelStatus.NerVersion,
			TopicVersion:     modelStatus.TopicVersion,
		},
		Details: map[string]string{
			"version": "1.0.0",
			"uptime":  "running",
		},
	}, nil
}

func (h *NLPGRPCHandler) GetProcessingStatus(ctx context.Context, req *emptypb.Empty) (*nlpv1.ProcessingStatusResponse, error) {
	logrus.Info("GetProcessingStatus gRPC method called")

	modelStatus := h.nlpService.GetModelStatus()

	return &nlpv1.ProcessingStatusResponse{
		Status:              "healthy",
		PendingArticles:     0, // Would be calculated from actual data
		ProcessingArticles:  0, // Would be calculated from actual data
		CompletedToday:      0, // Would be calculated from actual data
		FailedToday:         0, // Would be calculated from actual data
		AvgProcessingTimeMs: 0, // Would be calculated from actual data
		LastProcessed:       timestamppb.Now(),
		Performance: &nlpv1.ModelPerformance{
			SentimentAccuracy: 0.85, // Mock data
			NerPrecision:      0.90, // Mock data
			TopicAccuracy:     0.82, // Mock data
			TotalProcessed:    0,    // Would be calculated from actual data
			LastUpdated:       timestamppb.Now(),
		},
	}, nil
}

// Helper conversion functions
func (h *NLPGRPCHandler) protoToModelArticle(proto *nlpv1.Article) *model.Article {
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

func (h *NLPGRPCHandler) modelToProtoAnalysisResult(result *model.AnalysisResult) *nlpv1.AnalysisResult {
	proto := &nlpv1.AnalysisResult{
		ArticleId:        result.ArticleID,
		ProcessingTimeMs: result.ProcessingTimeMs,
		Status:           result.Status,
		ErrorMessage:     result.ErrorMessage,
		CreatedAt:        timestamppb.New(result.CreatedAt),
	}

	// Add sentiment analysis if present
	if result.SentimentAnalysis != nil {
		proto.Sentiment = &nlpv1.SentimentAnalysis{
			ArticleId:         result.SentimentAnalysis.ArticleID,
			CompoundScore:     result.SentimentAnalysis.CompoundScore,
			Confidence:        result.SentimentAnalysis.Confidence,
			PrimarySymbol:     result.SentimentAnalysis.PrimarySymbol,
			ModelVersion:      result.SentimentAnalysis.ModelVersion,
			AnalysisTimestamp: timestamppb.New(result.SentimentAnalysis.AnalysisTimestamp),
		}
	}

	// Add entity recognition if present
	if len(result.EntityRecognition) > 0 {
		proto.Entities = make([]*nlpv1.EntityRecognition, len(result.EntityRecognition))
		for i, entity := range result.EntityRecognition {
			proto.Entities[i] = &nlpv1.EntityRecognition{
				ArticleId:      entity.ArticleID,
				EntityText:     entity.EntityText,
				EntityType:     entity.EntityType,
				StockSymbol:    entity.StockSymbol,
				Confidence:     entity.Confidence,
				EntityCategory: entity.EntityCategory,
			}
		}
	}

	// Add topic classification if present
	if result.TopicClassification != nil {
		proto.Topic = &nlpv1.TopicClassification{
			ArticleId:              result.TopicClassification.ArticleID,
			PrimaryTopic:           result.TopicClassification.PrimaryTopic,
			PrimaryTopicConfidence: result.TopicClassification.PrimaryTopicConfidence,
			Keywords:               result.TopicClassification.Keywords,
			UrgencyScore:           result.TopicClassification.UrgencyScore,
			BreakingNewsIndicator:  result.TopicClassification.BreakingNewsIndicator == 1,
		}
	}

	return proto
}

func (h *NLPGRPCHandler) protoToModelProcessingOptions(proto *nlpv1.ProcessingOptions) *model.ProcessingOptions {
	if proto == nil {
		return &model.ProcessingOptions{
			EnableSentiment:     true,
			EnableNER:           true,
			EnableTopic:         true,
			EnableKeywords:      true,
			ConfidenceThreshold: 0.7,
			ModelVersion:        "1.0",
		}
	}

	return &model.ProcessingOptions{
		EnableSentiment:     proto.EnableSentiment,
		EnableNER:           proto.EnableNer,
		EnableTopic:         proto.EnableTopic,
		EnableKeywords:      proto.EnableKeywords,
		ConfidenceThreshold: proto.ConfidenceThreshold,
		ModelVersion:        proto.ModelVersion,
	}
}
