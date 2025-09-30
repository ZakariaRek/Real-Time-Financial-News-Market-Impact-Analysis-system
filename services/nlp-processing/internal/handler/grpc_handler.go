// services/nlp-processing/internal/handler/grpc_handler.go
package handler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/repository"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/service"
	newsv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/proto/gen/news/v1"
)

// GRPCHandler implementation
type GRPCHandler struct {
	newsv1.UnimplementedNewsServiceServer
	ingestionService service.IngestionService
	articleRepo      repository.ArticleRepository
	sourceRepo       repository.SourceRepository
	logRepo          repository.ProcessingLogRepository
}

func NewGRPCHandler(
	ingestionService service.IngestionService,
	articleRepo repository.ArticleRepository,
	sourceRepo repository.SourceRepository,
	logRepo repository.ProcessingLogRepository,
) *GRPCHandler {
	return &GRPCHandler{
		ingestionService: ingestionService,
		articleRepo:      articleRepo,
		sourceRepo:       sourceRepo,
		logRepo:          logRepo,
	}
}

// Article operations
func (h *GRPCHandler) CreateArticle(ctx context.Context, req *newsv1.CreateArticleRequest) (*newsv1.CreateArticleResponse, error) {
	// Validate request
	if req.Title == "" || req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "title and url are required")
	}

	// Parse published_at
	publishedAt := time.Now().UTC()
	if req.PublishedAt != nil {
		publishedAt = req.PublishedAt.AsTime()
	}

	// Generate content hash
	hash := sha256.Sum256([]byte(req.Title + req.Content + req.Url))
	contentHash := fmt.Sprintf("%x", hash)

	article := &model.Article{
		SourceID:         uint(req.SourceId),
		Title:            req.Title,
		Content:          req.Content,
		URL:              req.Url,
		Symbols:          req.Symbols,
		PublishedAt:      publishedAt,
		ProcessingStatus: model.StatusPending,
		ContentHash:      contentHash,
	}

	if err := h.articleRepo.Create(ctx, article); err != nil {
		logrus.WithError(err).Error("Failed to create article via gRPC")
		return nil, status.Error(codes.Internal, "failed to create article")
	}

	return &newsv1.CreateArticleResponse{
		Article: h.modelToProtoArticle(article),
	}, nil
}

func (h *GRPCHandler) GetArticle(ctx context.Context, req *newsv1.GetArticleRequest) (*newsv1.GetArticleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid article ID format")
	}

	article, err := h.articleRepo.GetByID(ctx, id)
	if err != nil {
		logrus.WithError(err).Error("Failed to get article via gRPC")
		return nil, status.Error(codes.NotFound, "article not found")
	}

	return &newsv1.GetArticleResponse{
		Article: h.modelToProtoArticle(article),
	}, nil
}

func (h *GRPCHandler) ListArticles(ctx context.Context, req *newsv1.ListArticlesRequest) (*newsv1.ListArticlesResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var articles []*model.Article
	var err error

	// Handle different query types
	if len(req.Symbols) > 0 {
		articles, err = h.articleRepo.GetBySymbols(ctx, req.Symbols, limit)
	} else if req.StartDate != nil && req.EndDate != nil {
		start := req.StartDate.AsTime()
		end := req.EndDate.AsTime()
		articles, err = h.articleRepo.GetByDateRange(ctx, start, end)
	} else if req.Status == "pending" {
		articles, err = h.articleRepo.GetPendingArticles(ctx, limit)
	} else {
		// Default: get recent articles
		end := time.Now().UTC()
		start := end.Add(-24 * time.Hour)
		articles, err = h.articleRepo.GetByDateRange(ctx, start, end)
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to list articles via gRPC")
		return nil, status.Error(codes.Internal, "failed to list articles")
	}

	// Convert to proto
	protoArticles := make([]*newsv1.Article, len(articles))
	for i, article := range articles {
		protoArticles[i] = h.modelToProtoArticle(article)
	}

	return &newsv1.ListArticlesResponse{
		Articles:   protoArticles,
		TotalCount: int32(len(articles)),
	}, nil
}

func (h *GRPCHandler) UpdateArticleStatus(ctx context.Context, req *newsv1.UpdateArticleStatusRequest) (*newsv1.UpdateArticleStatusResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid article ID format")
	}

	processingStatus := model.ProcessingStatus(req.Status)
	if err := h.articleRepo.UpdateStatus(ctx, id, processingStatus); err != nil {
		logrus.WithError(err).Error("Failed to update article status via gRPC")
		return nil, status.Error(codes.Internal, "failed to update article status")
	}

	return &newsv1.UpdateArticleStatusResponse{
		Success: true,
		Message: "Status updated successfully",
	}, nil
}

func (h *GRPCHandler) DeleteArticle(ctx context.Context, req *newsv1.DeleteArticleRequest) (*newsv1.DeleteArticleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid article ID format")
	}

	if err := h.articleRepo.Delete(ctx, id); err != nil {
		logrus.WithError(err).Error("Failed to delete article via gRPC")
		return nil, status.Error(codes.Internal, "failed to delete article")
	}

	return &newsv1.DeleteArticleResponse{
		Success: true,
		Message: "Article deleted successfully",
	}, nil
}

// Source operations
func (h *GRPCHandler) CreateSource(ctx context.Context, req *newsv1.CreateSourceRequest) (*newsv1.CreateSourceResponse, error) {
	if req.Name == "" || req.SourceType == "" {
		return nil, status.Error(codes.InvalidArgument, "name and source_type are required")
	}

	source := &model.NewsSource{
		Name:               req.Name,
		SourceType:         req.SourceType,
		BaseURL:            req.BaseUrl,
		RateLimitPerMinute: int(req.RateLimitPerMinute),
		Status:             req.Status,
	}

	if source.Status == "" {
		source.Status = "active"
	}

	if err := h.sourceRepo.Create(ctx, source); err != nil {
		logrus.WithError(err).Error("Failed to create source via gRPC")
		return nil, status.Error(codes.Internal, "failed to create source")
	}

	return &newsv1.CreateSourceResponse{
		Source: h.modelToProtoSource(source),
	}, nil
}

func (h *GRPCHandler) GetSource(ctx context.Context, req *newsv1.GetSourceRequest) (*newsv1.GetSourceResponse, error) {
	source, err := h.sourceRepo.GetByID(ctx, uint(req.Id))
	if err != nil {
		logrus.WithError(err).Error("Failed to get source via gRPC")
		return nil, status.Error(codes.NotFound, "source not found")
	}

	return &newsv1.GetSourceResponse{
		Source: h.modelToProtoSource(source),
	}, nil
}

func (h *GRPCHandler) ListSources(ctx context.Context, req *newsv1.ListSourcesRequest) (*newsv1.ListSourcesResponse, error) {
	var sources []*model.NewsSource
	var err error

	if req.ActiveOnly {
		sources, err = h.sourceRepo.GetActive(ctx)
	} else {
		sources, err = h.sourceRepo.GetActive(ctx) // For now, we only have GetActive method
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to list sources via gRPC")
		return nil, status.Error(codes.Internal, "failed to list sources")
	}

	// Convert to proto
	protoSources := make([]*newsv1.NewsSource, len(sources))
	for i, source := range sources {
		protoSources[i] = h.modelToProtoSource(source)
	}

	return &newsv1.ListSourcesResponse{
		Sources:    protoSources,
		TotalCount: int32(len(sources)),
	}, nil
}

func (h *GRPCHandler) UpdateSource(ctx context.Context, req *newsv1.UpdateSourceRequest) (*newsv1.UpdateSourceResponse, error) {
	source := &model.NewsSource{
		ID:                 uint(req.Id),
		Name:               req.Name,
		SourceType:         req.SourceType,
		BaseURL:            req.BaseUrl,
		RateLimitPerMinute: int(req.RateLimitPerMinute),
		Status:             req.Status,
	}

	if err := h.sourceRepo.Update(ctx, source); err != nil {
		logrus.WithError(err).Error("Failed to update source via gRPC")
		return nil, status.Error(codes.Internal, "failed to update source")
	}

	return &newsv1.UpdateSourceResponse{
		Source: h.modelToProtoSource(source),
	}, nil
}

func (h *GRPCHandler) DeleteSource(ctx context.Context, req *newsv1.DeleteSourceRequest) (*newsv1.DeleteSourceResponse, error) {
	if err := h.sourceRepo.Delete(ctx, uint(req.Id)); err != nil {
		logrus.WithError(err).Error("Failed to delete source via gRPC")
		return nil, status.Error(codes.Internal, "failed to delete source")
	}

	return &newsv1.DeleteSourceResponse{
		Success: true,
		Message: "Source deleted successfully",
	}, nil
}

// Ingestion operations
func (h *GRPCHandler) TriggerManualIngestion(ctx context.Context, req *newsv1.TriggerManualIngestionRequest) (*newsv1.TriggerManualIngestionResponse, error) {
	var err error
	var articlesIngested int32 = 0

	switch req.SourceType {
	case "rss":
		err = h.ingestionService.IngestFromRSS(ctx)
	case "newsapi":
		err = h.ingestionService.IngestFromNewsAPI(ctx)
	case "twitter":
		return nil, status.Error(codes.Unimplemented, "Twitter ingestion not implemented yet")
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid source type")
	}

	if err != nil {
		logrus.WithError(err).Error("Manual ingestion failed via gRPC")
		return nil, status.Error(codes.Internal, "ingestion failed")
	}

	return &newsv1.TriggerManualIngestionResponse{
		Success:          true,
		Message:          "Ingestion completed successfully",
		ArticlesIngested: articlesIngested,
		Timestamp:        timestamppb.Now(),
	}, nil
}

func (h *GRPCHandler) GetIngestionStatus(ctx context.Context, req *emptypb.Empty) (*newsv1.GetIngestionStatusResponse, error) {
	// Get pending articles count
	pendingArticles, err := h.articleRepo.GetPendingArticles(ctx, 1000)
	if err != nil {
		logrus.WithError(err).Error("Failed to get pending articles count")
		return nil, status.Error(codes.Internal, "failed to get ingestion status")
	}

	// Get active sources
	sources, err := h.sourceRepo.GetActive(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get active sources")
		return nil, status.Error(codes.Internal, "failed to get source statuses")
	}

	// Convert sources to status
	sourceStatuses := make([]*newsv1.SourceStatus, len(sources))
	for i, source := range sources {
		sourceStatuses[i] = &newsv1.SourceStatus{
			SourceId:      uint32(source.ID),
			SourceName:    source.Name,
			Status:        source.Status,
			LastFetch:     timestamppb.New(source.UpdatedAt),
			SuccessRate:   source.SuccessRate,
			ArticlesToday: 0, // This would need to be calculated from actual data
		}
	}

	return &newsv1.GetIngestionStatusResponse{
		Status:             "healthy",
		PendingArticles:    int32(len(pendingArticles)),
		ProcessingArticles: 0, // This would need to be calculated
		FailedArticles:     0, // This would need to be calculated
		LastIngestion:      timestamppb.Now(),
		SourceStatuses:     sourceStatuses,
	}, nil
}

func (h *GRPCHandler) GetProcessingLogs(ctx context.Context, req *newsv1.GetProcessingLogsRequest) (*newsv1.GetProcessingLogsResponse, error) {
	if req.ArticleId == "" {
		return nil, status.Error(codes.InvalidArgument, "article_id is required")
	}

	articleID, err := uuid.Parse(req.ArticleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid article ID format")
	}

	logs, err := h.logRepo.GetByArticleID(ctx, articleID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get processing logs via gRPC")
		return nil, status.Error(codes.Internal, "failed to get processing logs")
	}

	// Convert to proto
	protoLogs := make([]*newsv1.ProcessingLog, len(logs))
	for i, log := range logs {
		protoLogs[i] = &newsv1.ProcessingLog{
			Id:               uint32(log.ID),
			ArticleId:        log.ArticleID.String(),
			ProcessingStage:  log.ProcessingStage,
			Status:           log.Status,
			ProcessingTimeMs: int32(log.ProcessingTimeMs),
			ErrorMessage:     log.ErrorMessage,
			CreatedAt:        timestamppb.New(log.CreatedAt),
		}
	}

	return &newsv1.GetProcessingLogsResponse{
		Logs:       protoLogs,
		TotalCount: int32(len(logs)),
	}, nil
}

func (h *GRPCHandler) HealthCheck(ctx context.Context, req *emptypb.Empty) (*newsv1.HealthCheckResponse, error) {
	return &newsv1.HealthCheckResponse{
		Status:         "healthy",
		Service:        "news-ingestion",
		Timestamp:      timestamppb.Now(),
		DatabaseStatus: "connected",
		Details: map[string]string{
			"version": "1.0.0",
			"uptime":  "running",
		},
	}, nil
}

// Helper functions to convert between model and proto
func (h *GRPCHandler) modelToProtoArticle(article *model.Article) *newsv1.Article {
	proto := &newsv1.Article{
		Id:               article.ID.String(),
		SourceId:         uint32(article.SourceID),
		Title:            article.Title,
		Content:          article.Content,
		Url:              article.URL,
		Symbols:          article.Symbols,
		PublishedAt:      timestamppb.New(article.PublishedAt),
		ProcessingStatus: string(article.ProcessingStatus),
		RelevanceScore:   article.RelevanceScore,
		ContentHash:      article.ContentHash,
		CreatedAt:        timestamppb.New(article.CreatedAt),
		UpdatedAt:        timestamppb.New(article.UpdatedAt),
	}

	// Add source if loaded
	if article.Source.ID != 0 {
		proto.Source = h.modelToProtoSource(&article.Source)
	}

	return proto
}

func (h *GRPCHandler) modelToProtoSource(source *model.NewsSource) *newsv1.NewsSource {
	return &newsv1.NewsSource{
		Id:                 uint32(source.ID),
		Name:               source.Name,
		SourceType:         source.SourceType,
		BaseUrl:            source.BaseURL,
		RateLimitPerMinute: int32(source.RateLimitPerMinute),
		Status:             source.Status,
		SuccessRate:        source.SuccessRate,
		CreatedAt:          timestamppb.New(source.CreatedAt),
		UpdatedAt:          timestamppb.New(source.UpdatedAt),
	}
}
