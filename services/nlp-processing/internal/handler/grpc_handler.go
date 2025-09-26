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
)

// Placeholder proto types - In production, these would be generated from .proto files
// Once you generate the proto files, replace these with the actual generated types

// Service interface
type NewsServiceServer interface {
	CreateArticle(context.Context, *CreateArticleRequest) (*CreateArticleResponse, error)
	GetArticle(context.Context, *GetArticleRequest) (*GetArticleResponse, error)
	ListArticles(context.Context, *ListArticlesRequest) (*ListArticlesResponse, error)
	UpdateArticleStatus(context.Context, *UpdateArticleStatusRequest) (*UpdateArticleStatusResponse, error)
	DeleteArticle(context.Context, *DeleteArticleRequest) (*DeleteArticleResponse, error)
	CreateSource(context.Context, *CreateSourceRequest) (*CreateSourceResponse, error)
	GetSource(context.Context, *GetSourceRequest) (*GetSourceResponse, error)
	ListSources(context.Context, *ListSourcesRequest) (*ListSourcesResponse, error)
	UpdateSource(context.Context, *UpdateSourceRequest) (*UpdateSourceResponse, error)
	DeleteSource(context.Context, *DeleteSourceRequest) (*DeleteSourceResponse, error)
	TriggerManualIngestion(context.Context, *TriggerManualIngestionRequest) (*TriggerManualIngestionResponse, error)
	GetIngestionStatus(context.Context, *emptypb.Empty) (*GetIngestionStatusResponse, error)
	GetProcessingLogs(context.Context, *GetProcessingLogsRequest) (*GetProcessingLogsResponse, error)
	HealthCheck(context.Context, *emptypb.Empty) (*HealthCheckResponse, error)
}

type UnimplementedNewsServiceServer struct{}

// Article-related proto messages
type Article struct {
	Id               string
	SourceId         uint32
	Title            string
	Content          string
	Url              string
	Symbols          []string
	PublishedAt      *timestamppb.Timestamp
	ProcessingStatus string
	RelevanceScore   float64
	ContentHash      string
	CreatedAt        *timestamppb.Timestamp
	UpdatedAt        *timestamppb.Timestamp
	Source           *NewsSource
}

type CreateArticleRequest struct {
	SourceId    uint32
	Title       string
	Content     string
	Url         string
	Symbols     []string
	PublishedAt *timestamppb.Timestamp
}

type CreateArticleResponse struct {
	Article *Article
}

type GetArticleRequest struct {
	Id string
}

type GetArticleResponse struct {
	Article *Article
}

type ListArticlesRequest struct {
	Limit     int32
	Symbols   []string
	StartDate *timestamppb.Timestamp
	EndDate   *timestamppb.Timestamp
	Status    string
}

type ListArticlesResponse struct {
	Articles   []*Article
	TotalCount int32
}

type UpdateArticleStatusRequest struct {
	Id     string
	Status string
}

type UpdateArticleStatusResponse struct {
	Success bool
	Message string
}

type DeleteArticleRequest struct {
	Id string
}

type DeleteArticleResponse struct {
	Success bool
	Message string
}

// Source-related proto messages
type NewsSource struct {
	Id                 uint32
	Name               string
	SourceType         string
	BaseUrl            string
	RateLimitPerMinute int32
	Status             string
	SuccessRate        float64
	CreatedAt          *timestamppb.Timestamp
	UpdatedAt          *timestamppb.Timestamp
}

type CreateSourceRequest struct {
	Name               string
	SourceType         string
	BaseUrl            string
	RateLimitPerMinute int32
	Status             string
}

type CreateSourceResponse struct {
	Source *NewsSource
}

type GetSourceRequest struct {
	Id uint32
}

type GetSourceResponse struct {
	Source *NewsSource
}

type ListSourcesRequest struct {
	ActiveOnly bool
}

type ListSourcesResponse struct {
	Sources    []*NewsSource
	TotalCount int32
}

type UpdateSourceRequest struct {
	Id                 uint32
	Name               string
	SourceType         string
	BaseUrl            string
	RateLimitPerMinute int32
	Status             string
}

type UpdateSourceResponse struct {
	Source *NewsSource
}

type DeleteSourceRequest struct {
	Id uint32
}

type DeleteSourceResponse struct {
	Success bool
	Message string
}

// Ingestion-related proto messages
type TriggerManualIngestionRequest struct {
	SourceType string
}

type TriggerManualIngestionResponse struct {
	Success          bool
	Message          string
	ArticlesIngested int32
	Timestamp        *timestamppb.Timestamp
}

type GetIngestionStatusResponse struct {
	Status             string
	PendingArticles    int32
	ProcessingArticles int32
	FailedArticles     int32
	LastIngestion      *timestamppb.Timestamp
	SourceStatuses     []*SourceStatus
}

type SourceStatus struct {
	SourceId      uint32
	SourceName    string
	Status        string
	LastFetch     *timestamppb.Timestamp
	SuccessRate   float64
	ArticlesToday int32
}

// Processing logs
type GetProcessingLogsRequest struct {
	ArticleId string
}

type GetProcessingLogsResponse struct {
	Logs       []*ProcessingLog
	TotalCount int32
}

type ProcessingLog struct {
	Id               uint32
	ArticleId        string
	ProcessingStage  string
	Status           string
	ProcessingTimeMs int32
	ErrorMessage     string
	CreatedAt        *timestamppb.Timestamp
}

// Health check
type HealthCheckResponse struct {
	Status         string
	Service        string
	Timestamp      *timestamppb.Timestamp
	DatabaseStatus string
	Details        map[string]string
	ModelStatus    *ModelStatus
}

// GRPCHandler implementation
type GRPCHandler struct {
	UnimplementedNewsServiceServer
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
func (h *GRPCHandler) CreateArticle(ctx context.Context, req *CreateArticleRequest) (*CreateArticleResponse, error) {
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

	return &CreateArticleResponse{
		Article: h.modelToProtoArticle(article),
	}, nil
}

func (h *GRPCHandler) GetArticle(ctx context.Context, req *GetArticleRequest) (*GetArticleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid article ID format")
	}

	article, err := h.articleRepo.GetByID(ctx, id)
	if err != nil {
		logrus.WithError(err).Error("Failed to get article via gRPC")
		return nil, status.Error(codes.NotFound, "article not found")
	}

	return &GetArticleResponse{
		Article: h.modelToProtoArticle(article),
	}, nil
}

func (h *GRPCHandler) ListArticles(ctx context.Context, req *ListArticlesRequest) (*ListArticlesResponse, error) {
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
	protoArticles := make([]*Article, len(articles))
	for i, article := range articles {
		protoArticles[i] = h.modelToProtoArticle(article)
	}

	return &ListArticlesResponse{
		Articles:   protoArticles,
		TotalCount: int32(len(articles)),
	}, nil
}

func (h *GRPCHandler) UpdateArticleStatus(ctx context.Context, req *UpdateArticleStatusRequest) (*UpdateArticleStatusResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid article ID format")
	}

	processingStatus := model.ProcessingStatus(req.Status)
	if err := h.articleRepo.UpdateStatus(ctx, id, processingStatus); err != nil {
		logrus.WithError(err).Error("Failed to update article status via gRPC")
		return nil, status.Error(codes.Internal, "failed to update article status")
	}

	return &UpdateArticleStatusResponse{
		Success: true,
		Message: "Status updated successfully",
	}, nil
}

func (h *GRPCHandler) DeleteArticle(ctx context.Context, req *DeleteArticleRequest) (*DeleteArticleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid article ID format")
	}

	if err := h.articleRepo.Delete(ctx, id); err != nil {
		logrus.WithError(err).Error("Failed to delete article via gRPC")
		return nil, status.Error(codes.Internal, "failed to delete article")
	}

	return &DeleteArticleResponse{
		Success: true,
		Message: "Article deleted successfully",
	}, nil
}

// Source operations
func (h *GRPCHandler) CreateSource(ctx context.Context, req *CreateSourceRequest) (*CreateSourceResponse, error) {
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

	return &CreateSourceResponse{
		Source: h.modelToProtoSource(source),
	}, nil
}

func (h *GRPCHandler) GetSource(ctx context.Context, req *GetSourceRequest) (*GetSourceResponse, error) {
	source, err := h.sourceRepo.GetByID(ctx, uint(req.Id))
	if err != nil {
		logrus.WithError(err).Error("Failed to get source via gRPC")
		return nil, status.Error(codes.NotFound, "source not found")
	}

	return &GetSourceResponse{
		Source: h.modelToProtoSource(source),
	}, nil
}

func (h *GRPCHandler) ListSources(ctx context.Context, req *ListSourcesRequest) (*ListSourcesResponse, error) {
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
	protoSources := make([]*NewsSource, len(sources))
	for i, source := range sources {
		protoSources[i] = h.modelToProtoSource(source)
	}

	return &ListSourcesResponse{
		Sources:    protoSources,
		TotalCount: int32(len(sources)),
	}, nil
}

func (h *GRPCHandler) UpdateSource(ctx context.Context, req *UpdateSourceRequest) (*UpdateSourceResponse, error) {
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

	return &UpdateSourceResponse{
		Source: h.modelToProtoSource(source),
	}, nil
}

func (h *GRPCHandler) DeleteSource(ctx context.Context, req *DeleteSourceRequest) (*DeleteSourceResponse, error) {
	if err := h.sourceRepo.Delete(ctx, uint(req.Id)); err != nil {
		logrus.WithError(err).Error("Failed to delete source via gRPC")
		return nil, status.Error(codes.Internal, "failed to delete source")
	}

	return &DeleteSourceResponse{
		Success: true,
		Message: "Source deleted successfully",
	}, nil
}

// Ingestion operations
func (h *GRPCHandler) TriggerManualIngestion(ctx context.Context, req *TriggerManualIngestionRequest) (*TriggerManualIngestionResponse, error) {
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

	return &TriggerManualIngestionResponse{
		Success:          true,
		Message:          "Ingestion completed successfully",
		ArticlesIngested: articlesIngested,
		Timestamp:        timestamppb.Now(),
	}, nil
}

func (h *GRPCHandler) GetIngestionStatus(ctx context.Context, req *emptypb.Empty) (*GetIngestionStatusResponse, error) {
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
	sourceStatuses := make([]*SourceStatus, len(sources))
	for i, source := range sources {
		sourceStatuses[i] = &SourceStatus{
			SourceId:      uint32(source.ID),
			SourceName:    source.Name,
			Status:        source.Status,
			LastFetch:     timestamppb.New(source.UpdatedAt),
			SuccessRate:   source.SuccessRate,
			ArticlesToday: 0, // This would need to be calculated from actual data
		}
	}

	return &GetIngestionStatusResponse{
		Status:             "healthy",
		PendingArticles:    int32(len(pendingArticles)),
		ProcessingArticles: 0, // This would need to be calculated
		FailedArticles:     0, // This would need to be calculated
		LastIngestion:      timestamppb.Now(),
		SourceStatuses:     sourceStatuses,
	}, nil
}

func (h *GRPCHandler) GetProcessingLogs(ctx context.Context, req *GetProcessingLogsRequest) (*GetProcessingLogsResponse, error) {
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
	protoLogs := make([]*ProcessingLog, len(logs))
	for i, log := range logs {
		protoLogs[i] = &ProcessingLog{
			Id:               uint32(log.ID),
			ArticleId:        log.ArticleID.String(),
			ProcessingStage:  log.ProcessingStage,
			Status:           log.Status,
			ProcessingTimeMs: int32(log.ProcessingTimeMs),
			ErrorMessage:     log.ErrorMessage,
			CreatedAt:        timestamppb.New(log.CreatedAt),
		}
	}

	return &GetProcessingLogsResponse{
		Logs:       protoLogs,
		TotalCount: int32(len(logs)),
	}, nil
}

func (h *GRPCHandler) HealthCheck(ctx context.Context, req *emptypb.Empty) (*HealthCheckResponse, error) {
	return &HealthCheckResponse{
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
func (h *GRPCHandler) modelToProtoArticle(article *model.Article) *Article {
	proto := &Article{
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

func (h *GRPCHandler) modelToProtoSource(source *model.NewsSource) *NewsSource {
	return &NewsSource{
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
