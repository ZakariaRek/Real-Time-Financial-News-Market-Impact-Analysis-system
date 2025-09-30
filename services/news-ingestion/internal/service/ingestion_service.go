// services/news-ingestion/internal/service/ingestion_service.go
package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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

	newsAPIClient client.NewsAPIClient
	rssClient     client.RSSClient
	twitterClient client.TwitterClient

	rateLimitManager *RateLimitManager
	symbolExtractor  *SymbolExtractor
	contentValidator *ContentValidator
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
	newsAPIClient client.NewsAPIClient,
	rssClient client.RSSClient,
	twitterClient client.TwitterClient,
) IngestionService {
	// Initialize symbol extractor with common financial symbols
	symbols := map[string]bool{
		"AAPL": true, "GOOGL": true, "MSFT": true, "AMZN": true, "TSLA": true,
		"META": true, "NVDA": true, "AMD": true, "NFLX": true, "CRM": true,
		"BTC": true, "ETH": true, "SPY": true, "QQQ": true, "GLD": true,
		"JPM": true, "BAC": true, "WFC": true, "GS": true, "C": true,
		"XOM": true, "CVX": true, "PG": true, "JNJ": true, "V": true,
	}

	//return &ingestionService{
	//	articleRepo:       articleRepo,
	//	sourceRepo:        sourceRepo,
	//	processingLogRepo: processingLogRepo,
	//	rateLimitRepo:     rateLimitRepo,
	//	deduplicationSvc:  deduplicationSvc,
	//	newsAPIClient:     newsAPIClient,
	//	rssClient:         rssClient,
	//	twitterClient:     twitterClient,
	//	rateLimitManager:  &RateLimitManager{rateLimitRepo: rateLimitRepo},
	//	symbolExtractor:   &SymbolExtractor{financialSymbols: symbols},
	//	contentValidator:  &ContentValidator{minContentLength: 50, maxContentLength: 50000}, // Changed from 100 to 50
	//}
	// Line ~51 - Change minimum from 50 to 20
	return &ingestionService{
		articleRepo:       articleRepo,
		sourceRepo:        sourceRepo,
		processingLogRepo: processingLogRepo,
		rateLimitRepo:     rateLimitRepo,
		deduplicationSvc:  deduplicationSvc,
		newsAPIClient:     newsAPIClient,
		rssClient:         rssClient,
		twitterClient:     twitterClient,
		rateLimitManager:  &RateLimitManager{rateLimitRepo: rateLimitRepo},
		symbolExtractor:   &SymbolExtractor{financialSymbols: symbols},
		contentValidator:  &ContentValidator{minContentLength: 20, maxContentLength: 50000}, // Changed 50 → 20
	}
}
func (s *ingestionService) ingestFromSingleRSSSource(ctx context.Context, source *model.NewsSource) error {
	logrus.Infof("🔍 Fetching RSS feed from: %s (%s)", source.Name, source.BaseURL)

	// Check rate limits
	if !s.rateLimitManager.CanMakeRequest(ctx, source.ID, source.RateLimitPerMinute) {
		logrus.Warnf("⚠️ Rate limit exceeded for source: %s", source.Name)
		return nil
	}

	// Fetch RSS feed
	feeds, err := s.rssClient.FetchFeed(ctx, source.BaseURL)
	if err != nil {
		logrus.WithError(err).Errorf("❌ Failed to fetch RSS feed from %s", source.Name)
		return fmt.Errorf("failed to fetch RSS feed: %w", err)
	}

	logrus.Infof("📰 Fetched %d items from %s", len(feeds), source.Name)

	if len(feeds) == 0 {
		logrus.Warnf("⚠️ No items returned from RSS feed: %s", source.Name)
		return nil
	}

	var processedCount int
	var skippedDuplicate int
	var skippedValidation int

	for idx, item := range feeds {
		logrus.Debugf("Processing item %d/%d: %s", idx+1, len(feeds), item.Title)

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
			logrus.WithError(err).Warnf("⚠️ Article validation failed: %s (length: %d)", article.Title, len(article.Content))
			skippedValidation++
			continue
		}

		// Check for duplicates
		if s.deduplicationSvc.IsDuplicate(ctx, article) {
			logrus.Debugf("⏭️ Duplicate article detected: %s", article.Title)
			skippedDuplicate++
			continue
		}

		// Save article
		if err := s.articleRepo.Create(ctx, article); err != nil {
			logrus.WithError(err).Errorf("❌ Failed to save article: %s", article.Title)
			continue
		}

		processedCount++
		s.logProcessingStep(ctx, article.ID, "ingestion", "completed", "")
		logrus.Infof("✅ Saved article: %s (ID: %s)", article.Title, article.ID)
	}

	logrus.Infof("📊 RSS ingestion summary for %s: %d saved, %d duplicates, %d validation failures, %d total",
		source.Name, processedCount, skippedDuplicate, skippedValidation, len(feeds))

	return nil
}

func (s *ingestionService) IngestFromRSS(ctx context.Context) error {
	logrus.Info("Starting RSS ingestion")

	// Get all active RSS sources
	sources, err := s.sourceRepo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active sources: %w", err)
	}

	var rssSources []*model.NewsSource
	for _, source := range sources {
		if strings.ToUpper(source.SourceType) == "RSS" {
			rssSources = append(rssSources, source)
		}
	}

	if len(rssSources) == 0 {
		logrus.Info("No RSS sources configured")
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(rssSources))

	for _, source := range rssSources {
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
		logrus.WithError(err).Error("RSS ingestion error")
	}

	if len(errors) > 0 {
		return fmt.Errorf("RSS ingestion completed with errors: %s", strings.Join(errors, "; "))
	}

	logrus.Info("RSS ingestion completed successfully")
	return nil
}

//func (s *ingestionService) ingestFromSingleRSSSource(ctx context.Context, source *model.NewsSource) error {
//	// Check rate limits
//	if !s.rateLimitManager.CanMakeRequest(ctx, source.ID, source.RateLimitPerMinute) {
//		logrus.Warnf("Rate limit exceeded for source: %s", source.Name)
//		return nil
//	}
//
//	// Fetch RSS feed
//	feeds, err := s.rssClient.FetchFeed(ctx, source.BaseURL)
//	if err != nil {
//		return fmt.Errorf("failed to fetch RSS feed: %w", err)
//	}
//
//	var processedCount int
//	for _, item := range feeds {
//		article := &model.Article{
//			SourceID:         source.ID,
//			Title:            item.Title,
//			Content:          item.Description,
//			URL:              item.Link,
//			PublishedAt:      item.PubDate,
//			ProcessingStatus: model.StatusPending,
//			ContentHash:      s.generateContentHash(item.Title + item.Description),
//		}
//
//		// Validate and enrich article
//		if err := s.validateAndEnrichArticle(ctx, article); err != nil {
//			logrus.WithError(err).Warnf("Article validation failed: %s", article.Title)
//			continue
//		}
//
//		// Check for duplicates
//		if s.deduplicationSvc.IsDuplicate(ctx, article) {
//			logrus.Debugf("Duplicate article detected: %s", article.Title)
//			continue
//		}
//
//		// Save article
//		if err := s.articleRepo.Create(ctx, article); err != nil {
//			logrus.WithError(err).Errorf("Failed to save article: %s", article.Title)
//			continue
//		}
//
//		processedCount++
//		s.logProcessingStep(ctx, article.ID, "ingestion", "completed", "")
//	}
//
//	logrus.Infof("Processed %d articles from RSS source: %s", processedCount, source.Name)
//	return nil
//}

func (s *ingestionService) IngestFromNewsAPI(ctx context.Context) error {
	logrus.Info("Starting NewsAPI ingestion")

	// Get NewsAPI sources
	sources, err := s.sourceRepo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active sources: %w", err)
	}

	var apiSources []*model.NewsSource
	for _, source := range sources {
		if strings.ToUpper(source.SourceType) == "API" || strings.ToUpper(source.SourceType) == "NEWSAPI" {
			apiSources = append(apiSources, source)
		}
	}

	if len(apiSources) == 0 {
		logrus.Info("No NewsAPI sources configured")
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(apiSources))

	for _, source := range apiSources {
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
		logrus.WithError(err).Error("NewsAPI ingestion error")
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

	// Fetch articles from NewsAPI with financial keywords
	queries := []string{"financial", "stock market", "earnings", "nasdaq", "dow jones"}
	var allArticles []*client.NewsAPIArticle

	for _, query := range queries {
		articles, err := s.newsAPIClient.GetFinancialNews(ctx, query)
		if err != nil {
			logrus.WithError(err).Warnf("Failed to fetch NewsAPI articles for query: %s", query)
			continue
		}
		allArticles = append(allArticles, articles...)
	}

	var processedCount int
	seenURLs := make(map[string]bool) // Deduplicate within the same batch

	for _, apiArticle := range allArticles {
		// Skip duplicates within the same batch
		if seenURLs[apiArticle.URL] {
			continue
		}
		seenURLs[apiArticle.URL] = true

		article := &model.Article{
			SourceID:         source.ID,
			Title:            apiArticle.Title,
			Content:          apiArticle.Content,
			URL:              apiArticle.URL,
			PublishedAt:      apiArticle.PublishedAt,
			ProcessingStatus: model.StatusPending,
			ContentHash:      s.generateContentHash(apiArticle.Title + apiArticle.Content),
		}

		// Use description as content if main content is empty
		if article.Content == "" {
			article.Content = apiArticle.Description
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

	var twitterSources []*model.NewsSource
	for _, source := range sources {
		if strings.ToUpper(source.SourceType) == "TWITTER" {
			twitterSources = append(twitterSources, source)
		}
	}

	if len(twitterSources) == 0 {
		logrus.Info("No Twitter sources configured")
		return nil
	}

	for _, source := range twitterSources {
		// Check rate limits
		if !s.rateLimitManager.CanMakeRequest(ctx, source.ID, source.RateLimitPerMinute) {
			logrus.Warnf("Rate limit exceeded for source: %s", source.Name)
			continue
		}

		tweets, err := s.twitterClient.GetFinancialTweets(ctx, []string{"$AAPL", "$GOOGL", "$TSLA"})
		if err != nil {
			logrus.WithError(err).Warnf("Failed to fetch tweets for source: %s", source.Name)
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
	hasher := sha256.New()
	hasher.Write([]byte(content))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func (s *ingestionService) calculateRelevanceScore(article *model.Article) float64 {
	score := 0.0

	// Base score for having symbols
	if len(article.Symbols) > 0 {
		score += 0.5
	}

	// Check for financial keywords
	financialKeywords := []string{"earnings", "revenue", "profit", "loss", "market", "stock", "trading", "investment", "financial", "nasdaq", "dow", "s&p"}
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

// Update the logProcessingStep method in ingestion_service.go

func (s *ingestionService) logProcessingStep(ctx context.Context, articleID interface{}, stage, status, errorMsg string) {
	// Convert articleID to UUID
	var articleUUID uuid.UUID
	var err error

	switch v := articleID.(type) {
	case uuid.UUID:
		articleUUID = v
	case string:
		articleUUID, err = uuid.Parse(v)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to parse article ID as UUID: %v", v)
			return
		}
	case fmt.Stringer:
		articleUUID, err = uuid.Parse(v.String())
		if err != nil {
			logrus.WithError(err).Errorf("Failed to parse article ID as UUID: %v", v.String())
			return
		}
	default:
		logrus.Errorf("Invalid article ID type: %T", v)
		return
	}

	log := &model.ArticleProcessingLog{
		ArticleID:       articleUUID,
		ProcessingStage: stage,
		Status:          status,
		ErrorMessage:    errorMsg,
	}

	if err := s.processingLogRepo.Create(ctx, log); err != nil {
		logrus.WithError(err).Error("Failed to log processing step")
	}
}

// Also need to import uuid at the top of the file:
// import "github.com/google/uuid"

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
		// Clean the word of punctuation
		cleanWord := strings.Trim(word, ".,!?;:")

		// Check for ticker symbols with $ prefix
		if strings.HasPrefix(cleanWord, "$") {
			symbol := strings.TrimPrefix(cleanWord, "$")
			if len(symbol) >= 1 && len(symbol) <= 5 {
				symbols = append(symbols, symbol)
			}
		}

		// Check against known symbols
		if se.financialSymbols[cleanWord] {
			symbols = append(symbols, cleanWord)
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
