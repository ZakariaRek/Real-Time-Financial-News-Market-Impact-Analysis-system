// services/nlp-processing/internal/handler/nlp_grpc_handler.go
package handler

import (
	"context"
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
)

// Since we don't have the generated proto files, we'll create placeholder interfaces
// In production, these would be generated from the .proto file

// Placeholder interfaces for the proto service
type NLPProcessingServiceServer interface {
	StreamArticles(stream NLPProcessingService_StreamArticlesServer) error
	ProcessArticle(context.Context, *ProcessArticleRequest) (*ProcessArticleResponse, error)
	ProcessBatch(context.Context, *BatchProcessRequest) (*BatchProcessResponse, error)
	GetAnalysisResult(context.Context, *GetAnalysisRequest) (*GetAnalysisResponse, error)
	GetSentimentTrends(context.Context, *SentimentTrendsRequest) (*SentimentTrendsResponse, error)
	GetEntityMentions(context.Context, *EntityMentionsRequest) (*EntityMentionsResponse, error)
	GetTopicDistribution(context.Context, *TopicDistributionRequest) (*TopicDistributionResponse, error)
	StreamSentimentUpdates(*SentimentStreamRequest, NLPProcessingService_StreamSentimentUpdatesServer) error
	StreamBreakingNews(*BreakingNewsRequest, NLPProcessingService_StreamBreakingNewsServer) error
	HealthCheck(context.Context, *emptypb.Empty) (*HealthCheckResponse, error)
	GetProcessingStatus(context.Context, *emptypb.Empty) (*ProcessingStatusResponse, error)
}

type UnimplementedNLPProcessingServiceServer struct{}

// Placeholder proto message types
//type Article struct {
//	Id          string
//	Title       string
//	Content     string
//	Url         string
//	Symbols     []string
//	PublishedAt *timestamppb.Timestamp
//	SourceId    uint32
//}

type ProcessArticleRequest struct {
	Article *Article
	Options *ProcessingOptions
}

type ProcessArticleResponse struct {
	Result  *AnalysisResult
	Success bool
	Message string
}

type BatchProcessRequest struct {
	Articles []*Article
	Options  *ProcessingOptions
}

type BatchProcessResponse struct {
	Results         []*AnalysisResult
	SuccessfulCount int32
	FailedCount     int32
	Errors          []string
}

type ProcessingOptions struct {
	EnableSentiment     bool
	EnableNer           bool
	EnableTopic         bool
	EnableKeywords      bool
	ConfidenceThreshold float32
	ModelVersion        string
}

type AnalysisResult struct {
	ArticleId        string
	Sentiment        *SentimentAnalysis
	Entities         []*EntityRecognition
	Topic            *TopicClassification
	ProcessingTimeMs int64
	Status           string
	ErrorMessage     string
	CreatedAt        *timestamppb.Timestamp
}

type SentimentAnalysis struct {
	ArticleId         string
	CompoundScore     float32
	Confidence        float32
	PrimarySymbol     string
	ModelVersion      string
	AnalysisTimestamp *timestamppb.Timestamp
}

type EntityRecognition struct {
	ArticleId      string
	EntityText     string
	EntityType     string
	StockSymbol    string
	Confidence     float32
	EntityCategory string
}

type TopicClassification struct {
	ArticleId              string
	PrimaryTopic           string
	PrimaryTopicConfidence float32
	Keywords               []string
	UrgencyScore           float32
	BreakingNewsIndicator  bool
}

type GetAnalysisRequest struct {
	ArticleId string
}

type GetAnalysisResponse struct {
	Result *AnalysisResult
	Found  bool
}

type SentimentTrendsRequest struct {
	Symbol    string
	StartTime *timestamppb.Timestamp
	EndTime   *timestamppb.Timestamp
	Interval  string
}

type SentimentTrendsResponse struct {
	Trends     []*SentimentTrend
	TotalCount int32
}

type SentimentTrend struct {
	Timestamp    *timestamppb.Timestamp
	Symbol       string
	AvgSentiment float64
	ArticleCount int64
	Volatility   float64
}

type EntityMentionsRequest struct {
	EntityText string
}

type EntityMentionsResponse struct {
	Mentions   []*EntityMention
	TotalCount int32
}

type EntityMention struct {
	ArticleId    string
	EntityText   string
	EntityType   string
	Confidence   float32
	MentionTime  *timestamppb.Timestamp
	ArticleTitle string
}

type TopicDistributionRequest struct {
	StartTime *timestamppb.Timestamp
	EndTime   *timestamppb.Timestamp
	Limit     int32
}

type TopicDistributionResponse struct {
	Topics []*TopicDistribution
}

type TopicDistribution struct {
	Topic         string
	ArticleCount  int64
	AvgConfidence float64
	Percentage    float64
}

type SentimentStreamRequest struct {
	Symbols   []string
	Threshold float32
}

type BreakingNewsRequest struct {
	Topics           []string
	UrgencyThreshold float32
}

type HealthCheckResponse1 struct {
	Status         string
	Service        string
	Timestamp      *timestamppb.Timestamp
	DatabaseStatus string
	ModelStatus1   *ModelStatus
	Details        map[string]string
}

type ModelStatus struct {
	FinbertLoaded    bool
	NerModelLoaded   bool
	TopicModelLoaded bool
	FinbertVersion   string
	NerVersion       string
	TopicVersion     string
}

type ProcessingStatusResponse struct {
	Status              string
	PendingArticles     int32
	ProcessingArticles  int32
	CompletedToday      int32
	FailedToday         int32
	AvgProcessingTimeMs float64
	LastProcessed       *timestamppb.Timestamp
	Performance         *ModelPerformance
}

type ModelPerformance struct {
	SentimentAccuracy float64
	NerPrecision      float64
	TopicAccuracy     float64
	TotalProcessed    int64
	LastUpdated       *timestamppb.Timestamp
}

// Stream interfaces (placeholders)
type NLPProcessingService_StreamArticlesServer interface {
	Send(*AnalysisResult) error
	Recv() (*Article, error)
}

type NLPProcessingService_StreamSentimentUpdatesServer interface {
	Send(*SentimentUpdate) error
}

type NLPProcessingService_StreamBreakingNewsServer interface {
	Send(*BreakingNewsAlert) error
}

type SentimentUpdate struct {
	Symbol            string
	CurrentSentiment  float32
	PreviousSentiment float32
	Change            float32
	ArticleCount      int32
	Timestamp         *timestamppb.Timestamp
}

type BreakingNewsAlert struct {
	ArticleId      string
	Title          string
	PrimaryTopic   string
	UrgencyScore   float32
	Entities       []string
	SentimentScore float32
	Timestamp      *timestamppb.Timestamp
}

type NLPGRPCHandler struct {
	UnimplementedNLPProcessingServiceServer
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

func (h *NLPGRPCHandler) StreamArticles(stream NLPProcessingService_StreamArticlesServer) error {
	logrus.Info("StreamArticles gRPC method called")

	articleChan := make(chan *model.Article, 100)
	resultChan := make(chan *model.AnalysisResult, 100)

	ctx := context.Background()

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

func (h *NLPGRPCHandler) ProcessArticle(ctx context.Context, req *ProcessArticleRequest) (*ProcessArticleResponse, error) {
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

	return &ProcessArticleResponse{
		Result:  h.modelToProtoAnalysisResult(result),
		Success: result.Status != "failed",
		Message: "Article processed successfully",
	}, nil
}

func (h *NLPGRPCHandler) ProcessBatch(ctx context.Context, req *BatchProcessRequest) (*BatchProcessResponse, error) {
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
	protoResults := make([]*AnalysisResult, len(results))
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

	return &BatchProcessResponse{
		Results:         protoResults,
		SuccessfulCount: successfulCount,
		FailedCount:     failedCount,
		Errors:          errors,
	}, nil
}

func (h *NLPGRPCHandler) GetAnalysisResult(ctx context.Context, req *GetAnalysisRequest) (*GetAnalysisResponse, error) {
	logrus.WithField("article_id", req.ArticleId).Info("GetAnalysisResult gRPC method called")

	if req.ArticleId == "" {
		return nil, status.Error(codes.InvalidArgument, "article_id is required")
	}

	result, err := h.nlpService.GetAnalysisResult(ctx, req.ArticleId)
	if err != nil {
		logrus.WithError(err).Error("Failed to get analysis result")
		return &GetAnalysisResponse{
			Result: nil,
			Found:  false,
		}, nil
	}

	return &GetAnalysisResponse{
		Result: h.modelToProtoAnalysisResult(result),
		Found:  true,
	}, nil
}

func (h *NLPGRPCHandler) GetSentimentTrends(ctx context.Context, req *SentimentTrendsRequest) (*SentimentTrendsResponse, error) {
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
	protoTrends := make([]*SentimentTrend, len(trends))
	for i, trend := range trends {
		protoTrends[i] = &SentimentTrend{
			Timestamp:    timestamppb.New(trend.Timestamp),
			Symbol:       trend.Symbol,
			AvgSentiment: trend.AvgSentiment,
			ArticleCount: trend.ArticleCount,
			Volatility:   trend.Volatility,
		}
	}

	return &SentimentTrendsResponse{
		Trends:     protoTrends,
		TotalCount: int32(len(trends)),
	}, nil
}

func (h *NLPGRPCHandler) GetEntityMentions(ctx context.Context, req *EntityMentionsRequest) (*EntityMentionsResponse, error) {
	logrus.WithField("entity", req.EntityText).Info("GetEntityMentions gRPC method called")

	// This would typically query the analysis repository for entity mentions
	// For now, return empty response
	return &EntityMentionsResponse{
		Mentions:   []*EntityMention{},
		TotalCount: 0,
	}, nil
}

func (h *NLPGRPCHandler) GetTopicDistribution(ctx context.Context, req *TopicDistributionRequest) (*TopicDistributionResponse, error) {
	logrus.Info("GetTopicDistribution gRPC method called")

	// This would typically query the analysis repository for topic distribution
	// For now, return stub data
	return &TopicDistributionResponse{
		Topics: []*TopicDistribution{},
	}, nil
}

func (h *NLPGRPCHandler) StreamSentimentUpdates(req *SentimentStreamRequest, stream NLPProcessingService_StreamSentimentUpdatesServer) error {
	logrus.Info("StreamSentimentUpdates gRPC method called")

	// This would implement real-time sentiment streaming
	// For now, return empty implementation
	return nil
}

func (h *NLPGRPCHandler) StreamBreakingNews(req *BreakingNewsRequest, stream NLPProcessingService_StreamBreakingNewsServer) error {
	logrus.Info("StreamBreakingNews gRPC method called")

	// This would implement real-time breaking news streaming
	// For now, return empty implementation
	return nil
}

func (h *NLPGRPCHandler) HealthCheck(ctx context.Context, req *emptypb.Empty) (*HealthCheckResponse, error) {
	logrus.Info("HealthCheck gRPC method called")

	modelStatus := h.nlpService.GetModelStatus()

	return &HealthCheckResponse{
		Status:         "healthy",
		Service:        "nlp-processing",
		Timestamp:      timestamppb.Now(),
		DatabaseStatus: "connected",
		ModelStatus: &ModelStatus{
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

func (h *NLPGRPCHandler) GetProcessingStatus(ctx context.Context, req *emptypb.Empty) (*ProcessingStatusResponse, error) {
	logrus.Info("GetProcessingStatus gRPC method called")

	return &ProcessingStatusResponse{
		Status:              "healthy",
		PendingArticles:     0, // Would be calculated from actual data
		ProcessingArticles:  0, // Would be calculated from actual data
		CompletedToday:      0, // Would be calculated from actual data
		FailedToday:         0, // Would be calculated from actual data
		AvgProcessingTimeMs: 0, // Would be calculated from actual data
		LastProcessed:       timestamppb.Now(),
		Performance: &ModelPerformance{
			SentimentAccuracy: 0.85, // Mock data
			NerPrecision:      0.90, // Mock data
			TopicAccuracy:     0.82, // Mock data
			TotalProcessed:    0,    // Would be calculated from actual data
			LastUpdated:       timestamppb.Now(),
		},
	}, nil
}

// Helper conversion functions
func (h *NLPGRPCHandler) protoToModelArticle(proto *Article) *model.Article {
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

func (h *NLPGRPCHandler) modelToProtoAnalysisResult(result *model.AnalysisResult) *AnalysisResult {
	proto := &AnalysisResult{
		ArticleId:        result.ArticleID,
		ProcessingTimeMs: result.ProcessingTimeMs,
		Status:           result.Status,
		ErrorMessage:     result.ErrorMessage,
		CreatedAt:        timestamppb.New(result.CreatedAt),
	}

	// Add sentiment analysis if present
	if result.SentimentAnalysis != nil {
		proto.Sentiment = &SentimentAnalysis{
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
		proto.Entities = make([]*EntityRecognition, len(result.EntityRecognition))
		for i, entity := range result.EntityRecognition {
			proto.Entities[i] = &EntityRecognition{
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
		proto.Topic = &TopicClassification{
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

func (h *NLPGRPCHandler) protoToModelProcessingOptions(proto *ProcessingOptions) *model.ProcessingOptions {
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
