package service

/*
import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/client"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/repository"
)

type IngestionService interface {
	IngestFromRSS(ctx context.Context) error
	IngestFromNewsAPI(ctx context.Context) error
	IngestFromTwitter(ctx context.Context) error
	ProcessArticle(ctx context.Context, article *model.Article) error
}

type ingestionService struct {
	articleRepo       repository.ArticleRepository
	sourceRepo        repository.SourceRepository
	processingLogRepo repository.ProcessingLogRepository
	rateLimitRepo     repository.RateLimitRepository
	deduplicationSvc  DeduplicationService
	scraperSvc        ScraperService
	newsAPIClient     client.NewsAPIClient
	rssClient         client.RSSClient
	twitterClient     client.TwitterClient
	rateLimitManager  *RateLimitManager
	symbolExtractor   *SymbolExtractor
	contentValidator  *ContentValidator
}

type RateLimitManager struct {
	rateLimitRepo repository.RateLimitRepository
}

type SymbolExtractor struct {
	financialSymbols map[string]bool
}

type ContentValidator struct {
	minContentLength int
	maxContentLength int
}

func NewIngestionService(
	articleRepo repository.ArticleRepository,
	sourceRepo repository.SourceRepository,
	processingLogRepo repository.ProcessingLogRepository,
	rateLimitRepo repository.RateLimitRepository,
	deduplicationSvc DeduplicationService,
	scraperSvc ScraperService,
	newsAPIClient client.NewsAPIClient,
	rssClient client.RSSClient,
	twitterClient client.TwitterClient,
) IngestionService {
	// Initialize symbol extractor with common financial symbols
	symbols := map[string]bool{
		"AAPL": true, "GOOGL": true, "MSFT": true, "AMZN": true, "TSLA": true,
		"META": true, "NVDA": true, "AMD": true, "NFLX": true, "CRM": true,
		"BTC": true, "ETH": true, "SPY": true, "QQQ": true, "GLD": true,
	}

	return &ingestionService{
		articleRepo:       articleRepo,
		sourceRepo:        sourceRepo,
		processingLogRepo: processingLogRepo,
		rateLimitRepo:     rateLimitRepo,
		deduplicationSvc:  deduplicationSvc,
		scraperSvc:        scraperSvc,
		newsAPIClient:     newsAPIClient,
		rssClient:         rssClient,
		twitterClient:     twitterClient,
		rateLimitManager:  &RateLimitManager{rateLimitRepo: rateLimitRepo},
		symbolExtractor:   &SymbolExtractor{financialSymbols: symbols},
		contentValidator:  &ContentValidator{minContentLength: 100, maxContentLength: 50000},
	}
}

func (s *ingestionService) IngestFromRSS(ctx context.Context) error {
	logrus.Info("Starting RSS ingestion")

	// Get all active RSS sources
	sources, err := s.sourceRepo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active sources: %w", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(sources))

	for _, source := range sources {
		if source.SourceType != "RSS" {
			continue
		}

		wg.Add(1)
		go func(src *model.NewsSource) {
			defer wg.Done()
			if err := s.ingestFromSingleRSSSource(ctx, src); err != nil {
				errChan <- fmt.Errorf("RSS ingestion failed for source %s: %w", src.Name, err)
			}
		}(source)
	}

	wg.Wait()
	close(errChan)

	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("RSS ingestion completed with errors: %s", strings.Join(errors, "; "))
	}

	logrus.Info("RSS ingestion completed successfully")
	return nil
}

func (s *ingestionService) ingestFromSingleRSSSource(ctx context.Context, source *model.NewsSource) error {
	// Check rate limits
	if !s.rateLimitManager.CanMakeRequest(ctx, source.ID, source.RateLimitPerMinute) {
		logrus.Warnf("Rate limit exceeded for source: %s", source.Name)
		return nil
	}

	// Fetch RSS feed
	feeds, err := s.rssClient.FetchFeed(ctx, source.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to fetch RSS feed: %w", err)
	}

	var processedCount int
	for _, item := range feeds {
		article := &model.Article{
			SourceID:         source.ID,
			Title:            item.Title,
			Content:          item.Description,
			URL:              item.Link,
			PublishedAt:      item.PubDate,
			ProcessingStatus: model.StatusPending,
			ContentHash:      s.generateContentHash(item.Title + item.Description),
		}

		// Validate and enrich article
		if err := s.validateAndEnrichArticle(ctx, article); err != nil {
			logrus.WithError(err).Warnf("Article validation failed: %s", article.Title)
			continue
		}

		// Check for duplicates
		if s.deduplicationSvc.IsDuplicate(ctx, article) {
			logrus.Debugf("Duplicate article detected: %s", article.Title)
			continue
		}

		// Save article
		if err := s.articleRepo.Create(ctx, article); err != nil {
			logrus.WithError(err).Errorf("Failed to save article: %s", article.Title)
			continue
		}

		processedCount++
		s.logProcessingStep(ctx, article.ID, "ingestion", "completed", "")
	}

	logrus.Infof("Processed %d articles from RSS source: %s", processedCount, source.Name)
	return nil
}

func (s *ingestionService) IngestFromNewsAPI(ctx context.Context) error {
	logrus.Info("Starting NewsAPI ingestion")

	// Get NewsAPI sources
	sources, err := s.sourceRepo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active sources: %w", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(sources))

	for _, source := range sources {
		if source.SourceType != "API" {
			continue
		}

		wg.Add(1)
		go func(src *model.NewsSource) {
			defer wg.Done()
			if err := s.ingestFromNewsAPISource(ctx, src); err != nil {
				errChan <- fmt.Errorf("NewsAPI ingestion failed for source %s: %w", src.Name, err)
			}
		}(source)
	}

	wg.Wait()
	close(errChan)

	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("NewsAPI ingestion completed with errors: %s", strings.Join(errors, "; "))
	}

	logrus.Info("NewsAPI ingestion completed successfully")
	return nil
}

func (s *ingestionService) ingestFromNewsAPISource(ctx context.Context, source *model.NewsSource) error {
	// Check rate limits
	if !s.rateLimitManager.CanMakeRequest(ctx, source.ID, source.RateLimitPerMinute) {
		logrus.Warnf("Rate limit exceeded for source: %s", source.Name)
		return nil
	}

	// Fetch articles from NewsAPI
	articles, err := s.newsAPIClient.GetFinancialNews(ctx, "financial")
	if err != nil {
		return fmt.Errorf("failed to fetch NewsAPI articles: %w", err)
	}

	var processedCount int
	for _, apiArticle := range articles {
		article := &model.Article{
			SourceID:         source.ID,
			Title:            apiArticle.Title,
			Content:          apiArticle.Content,
			URL:              apiArticle.URL,
			PublishedAt:      apiArticle.PublishedAt,
			ProcessingStatus: model.StatusPending,
			ContentHash:      s.generateContentHash(apiArticle.Title + apiArticle.Content),
		}

		// Validate and enrich article
		if err := s.validateAndEnrichArticle(ctx, article); err != nil {
			logrus.WithError(err).Warnf("Article validation failed: %s", article.Title)
			continue
		}

		// Check for duplicates
		if s.deduplicationSvc.IsDuplicate(ctx, article) {
			logrus.Debugf("Duplicate article detected: %s", article.Title)
			continue
		}

		// Save article
		if err := s.articleRepo.Create(ctx, article); err != nil {
			logrus.WithError(err).Errorf("Failed to save article: %s", article.Title)
			continue
		}

		processedCount++
		s.logProcessingStep(ctx, article.ID, "ingestion", "completed", "")
	}

	logrus.Infof("Processed %d articles from NewsAPI source: %s", processedCount, source.Name)
	return nil
}

func (s *ingestionService) IngestFromTwitter(ctx context.Context) error {
	logrus.Info("Starting Twitter ingestion")

	// Get Twitter sources
	sources, err := s.sourceRepo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active sources: %w", err)
	}

	for _, source := range sources {
		if source.SourceType != "TWITTER" {
			continue
		}

		// Check rate limits
		if !s.rateLimitManager.CanMakeRequest(ctx, source.ID, source.RateLimitPerMinute) {
			logrus.Warnf("Rate limit exceeded for source: %s", source.Name)
			continue
		}

		tweets, err := s.twitterClient.GetFinancialTweets(ctx, []string{"$AAPL", "$GOOGL", "$TSLA"})
		if err != nil {
			logrus.WithError(err).Errorf("Failed to fetch tweets for source: %s", source.Name)
			continue
		}

		var processedCount int
		for _, tweet := range tweets {
			article := &model.Article{
				SourceID:         source.ID,
				Title:            fmt.Sprintf("Tweet from @%s", tweet.Author),
				Content:          tweet.Text,
				URL:              tweet.URL,
				PublishedAt:      tweet.CreatedAt,
				ProcessingStatus: model.StatusPending,
				ContentHash:      s.generateContentHash(tweet.Text),
			}

			// Extract symbols from tweet
			article.Symbols = s.symbolExtractor.ExtractFromText(tweet.Text)

			// Validate content length for tweets
			if len(article.Content) < 50 {
				continue
			}

			// Check for duplicates
			if s.deduplicationSvc.IsDuplicate(ctx, article) {
				continue
			}

			// Save article
			if err := s.articleRepo.Create(ctx, article); err != nil {
				logrus.WithError(err).Errorf("Failed to save tweet: %s", article.Title)
				continue
			}

			processedCount++
		}

		logrus.Infof("Processed %d tweets from source: %s", processedCount, source.Name)
	}

	logrus.Info("Twitter ingestion completed successfully")
	return nil
}

func (s *ingestionService) ProcessArticle(ctx context.Context, article *model.Article) error {
	s.logProcessingStep(ctx, article.ID, "processing", "started", "")

	// Update status to processing
	if err := s.articleRepo.UpdateStatus(ctx, article.ID, model.StatusProcessing); err != nil {
		return fmt.Errorf("failed to update article status: %w", err)
	}

	// Validate and enrich the article
	if err := s.validateAndEnrichArticle(ctx, article); err != nil {
		s.logProcessingStep(ctx, article.ID, "processing", "failed", err.Error())
		s.articleRepo.UpdateStatus(ctx, article.ID, model.StatusFailed)
		return fmt.Errorf("article processing failed: %w", err)
	}

	// Update status to completed
	if err := s.articleRepo.UpdateStatus(ctx, article.ID, model.StatusCompleted); err != nil {
		return fmt.Errorf("failed to update article status: %w", err)
	}

	s.logProcessingStep(ctx, article.ID, "processing", "completed", "")
	return nil
}

func (s *ingestionService) validateAndEnrichArticle(ctx context.Context, article *model.Article) error {
	// Content validation
	if !s.contentValidator.IsValid(article.Content) {
		return fmt.Errorf("content validation failed")
	}

	// Extract financial symbols
	article.Symbols = s.symbolExtractor.ExtractFromText(article.Title + " " + article.Content)

	// Calculate relevance score (simplified)
	article.RelevanceScore = s.calculateRelevanceScore(article)

	return nil
}

func (s *ingestionService) generateContentHash(content string) string {
	hasher := md5.New()
	hasher.Write([]byte(content))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (s *ingestionService) calculateRelevanceScore(article *model.Article) float64 {
	score := 0.0

	// Base score for having symbols
	if len(article.Symbols) > 0 {
		score += 0.5
	}

	// Check for financial keywords
	financialKeywords := []string{"earnings", "revenue", "profit", "loss", "market", "stock", "trading", "investment"}
	content := strings.ToLower(article.Title + " " + article.Content)

	for _, keyword := range financialKeywords {
		if strings.Contains(content, keyword) {
			score += 0.1
		}
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (s *ingestionService) logProcessingStep(ctx context.Context, articleID interface{}, stage, status, errorMsg string) {
	log := &model.ArticleProcessingLog{
		ArticleID:       articleID.(interface{ String() string }).String(),
		ProcessingStage: stage,
		Status:          status,
		ErrorMessage:    errorMsg,
	}

	if err := s.processingLogRepo.Create(ctx, log); err != nil {
		logrus.WithError(err).Error("Failed to log processing step")
	}
}

// Rate limit manager methods
func (rlm *RateLimitManager) CanMakeRequest(ctx context.Context, sourceID uint, rateLimit int) bool {
	now := time.Now().UTC()
	timeWindow := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)

	currentCount, err := rlm.rateLimitRepo.GetCurrentCount(ctx, sourceID, timeWindow)
	if err != nil {
		logrus.WithError(err).Error("Failed to get rate limit count")
		return true // Allow request on error
	}

	if currentCount >= rateLimit {
		return false
	}

	// Increment count
	if err := rlm.rateLimitRepo.IncrementCount(ctx, sourceID, timeWindow); err != nil {
		logrus.WithError(err).Error("Failed to increment rate limit count")
	}

	return true
}

// Symbol extractor methods
func (se *SymbolExtractor) ExtractFromText(text string) []string {
	var symbols []string
	words := strings.Fields(strings.ToUpper(text))

	for _, word := range words {
		// Check for ticker symbols with $ prefix
		if strings.HasPrefix(word, "$") {
			symbol := strings.TrimPrefix(word, "$")
			if len(symbol) >= 1 && len(symbol) <= 5 {
				symbols = append(symbols, symbol)
			}
		}

		// Check against known symbols
		if se.financialSymbols[word] {
			symbols = append(symbols, word)
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueSymbols []string
	for _, symbol := range symbols {
		if !seen[symbol] {
			uniqueSymbols = append(uniqueSymbols, symbol)
			seen[symbol] = true
		}
	}

	return uniqueSymbols
}

// Content validator methods
func (cv *ContentValidator) IsValid(content string) bool {
	if len(content) < cv.minContentLength {
		return false
	}
	if len(content) > cv.maxContentLength {
		return false
	}

	// Check for minimum word count
	words := strings.Fields(content)
	if len(words) < 10 {
		return false
	}

	return true
}
*/
