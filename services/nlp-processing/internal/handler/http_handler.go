// services/news-ingestion/internal/handler/http_handler.go
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/repository"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type HTTPHandler struct {
	ingestionService service.IngestionService
	articleRepo      repository.ArticleRepository
	sourceRepo       repository.SourceRepository
}

func NewHTTPHandler(
	ingestionService service.IngestionService,
	articleRepo repository.ArticleRepository,
	sourceRepo repository.SourceRepository,
) *HTTPHandler {
	return &HTTPHandler{
		ingestionService: ingestionService,
		articleRepo:      articleRepo,
		sourceRepo:       sourceRepo,
	}
}

// Article handlers
func (h *HTTPHandler) CreateArticle(c *gin.Context) {
	var req struct {
		SourceID    uint     `json:"source_id" binding:"required"`
		Title       string   `json:"title" binding:"required"`
		Content     string   `json:"content"`
		URL         string   `json:"url" binding:"required"`
		Symbols     []string `json:"symbols"`
		PublishedAt string   `json:"published_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse published_at
	publishedAt := time.Now().UTC()
	if req.PublishedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, req.PublishedAt); err == nil {
			publishedAt = parsed
		}
	}

	article := &model.Article{
		SourceID:         req.SourceID,
		Title:            req.Title,
		Content:          req.Content,
		URL:              req.URL,
		Symbols:          req.Symbols,
		PublishedAt:      publishedAt,
		ProcessingStatus: model.StatusPending,
	}

	if err := h.articleRepo.Create(c.Request.Context(), article); err != nil {
		logrus.WithError(err).Error("Failed to create article")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article"})
		return
	}

	c.JSON(http.StatusCreated, article)
}

func (h *HTTPHandler) GetArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	article, err := h.articleRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		logrus.WithError(err).Error("Failed to get article")
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, article)
}

func (h *HTTPHandler) ListArticles(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit > 100 {
		limit = 100
	}

	status := c.Query("status")
	symbols := c.QueryArray("symbols")

	var articles []*model.Article
	var err error

	if len(symbols) > 0 {
		articles, err = h.articleRepo.GetBySymbols(c.Request.Context(), symbols, limit)
	} else if status != "" {
		// Get articles by status
		articles, err = h.articleRepo.GetPendingArticles(c.Request.Context(), limit)
	} else {
		// Get recent articles
		end := time.Now().UTC()
		start := end.Add(-24 * time.Hour)
		articles, err = h.articleRepo.GetByDateRange(c.Request.Context(), start, end)
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to list articles")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list articles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"articles": articles,
		"count":    len(articles),
	})
}

func (h *HTTPHandler) UpdateArticleStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := model.ProcessingStatus(req.Status)
	if err := h.articleRepo.UpdateStatus(c.Request.Context(), id, status); err != nil {
		logrus.WithError(err).Error("Failed to update article status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

// Source handlers
func (h *HTTPHandler) CreateSource(c *gin.Context) {
	var source model.NewsSource
	if err := c.ShouldBindJSON(&source); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.sourceRepo.Create(c.Request.Context(), &source); err != nil {
		logrus.WithError(err).Error("Failed to create source")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create source"})
		return
	}

	c.JSON(http.StatusCreated, source)
}

func (h *HTTPHandler) GetSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source ID"})
		return
	}

	source, err := h.sourceRepo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		logrus.WithError(err).Error("Failed to get source")
		c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
		return
	}

	c.JSON(http.StatusOK, source)
}

func (h *HTTPHandler) ListSources(c *gin.Context) {
	activeOnly := c.Query("active") == "true"

	var sources []*model.NewsSource
	var err error

	if activeOnly {
		sources, err = h.sourceRepo.GetActive(c.Request.Context())
	} else {
		// For simplicity, we'll get active sources. You can extend this to get all sources
		sources, err = h.sourceRepo.GetActive(c.Request.Context())
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to list sources")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list sources"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sources": sources,
		"count":   len(sources),
	})
}

func (h *HTTPHandler) UpdateSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source ID"})
		return
	}

	var updateReq model.NewsSource
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq.ID = uint(id)
	if err := h.sourceRepo.Update(c.Request.Context(), &updateReq); err != nil {
		logrus.WithError(err).Error("Failed to update source")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update source"})
		return
	}

	c.JSON(http.StatusOK, updateReq)
}

// Ingestion handlers
func (h *HTTPHandler) TriggerManualIngestion(c *gin.Context) {
	var req struct {
		SourceType string `json:"source_type" binding:"required"` // rss, newsapi, twitter
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	switch req.SourceType {
	case "rss":
		if h.ingestionService != nil {
			err = h.ingestionService.IngestFromRSS(c.Request.Context())
		} else {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Ingestion service not implemented yet"})
			return
		}
	case "newsapi":
		if h.ingestionService != nil {
			err = h.ingestionService.IngestFromNewsAPI(c.Request.Context())
		} else {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Ingestion service not implemented yet"})
			return
		}
	case "twitter":
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Twitter ingestion not implemented yet"})
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source type"})
		return
	}

	if err != nil {
		logrus.WithError(err).Error("Manual ingestion failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ingestion failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Ingestion completed successfully",
		"source_type": req.SourceType,
		"timestamp":   time.Now().UTC(),
	})
}

func (h *HTTPHandler) GetIngestionStatus(c *gin.Context) {
	// This is a simplified status endpoint
	// In a real implementation, you might track ingestion metrics

	pendingCount := 0
	if articles, err := h.articleRepo.GetPendingArticles(c.Request.Context(), 1000); err == nil {
		pendingCount = len(articles)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "healthy",
		"pending_articles": pendingCount,
		"last_check":       time.Now().UTC(),
	})
}
