// services/news-ingestion/internal/service/extraction_service.go
package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/client"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/repository"
)

// ExtractedArticle represents the fully extracted and enriched article
type ExtractedArticle struct {
	ID               uuid.UUID `json:"id"`
	SourceID         uint      `json:"source_id"`
	SourceName       string    `json:"source_name"`
	Title            string    `json:"title"`
	Content          string    `json:"content"`
	Summary          string    `json:"summary"`
	URL              string    `json:"url"`
	ImageURL         string    `json:"image_url"`
	Author           string    `json:"author"`
	Symbols          []string  `json:"symbols"`
	Keywords         []string  `json:"keywords"`
	PublishedAt      time.Time `json:"published_at"`
	ContentHash      string    `json:"content_hash"`
	RelevanceScore   float64   `json:"relevance_score"`
	ProcessingStatus string    `json:"processing_status"`
	Language         string    `json:"language"`
	Category         string    `json:"category"`
}

// DataExtractionService handles advanced data extraction and enrichment
type DataExtractionService struct {
	articleRepo        repository.ArticleRepository
	sourceRepo         repository.SourceRepository
	deduplicationSvc   DeduplicationService
	symbolExtractor    *SymbolExtractor
	keywordExtractor   *KeywordExtractor
	categoryClassifier *CategoryClassifier
}

// KeywordExtractor extracts relevant financial keywords
type KeywordExtractor struct {
	financialKeywords map[string]float64
}

// CategoryClassifier classifies articles into categories
type CategoryClassifier struct {
	categories map[string][]string
}

func NewDataExtractionService(
	articleRepo repository.ArticleRepository,
	sourceRepo repository.SourceRepository,
	deduplicationSvc DeduplicationService,
) *DataExtractionService {
	return &DataExtractionService{
		articleRepo:        articleRepo,
		sourceRepo:         sourceRepo,
		deduplicationSvc:   deduplicationSvc,
		symbolExtractor:    newSymbolExtractor(),
		keywordExtractor:   newKeywordExtractor(),
		categoryClassifier: newCategoryClassifier(),
	}
}

// ExtractFromNewsAPI extracts and processes articles from NewsAPI
func (s *DataExtractionService) ExtractFromNewsAPI(ctx context.Context, apiArticles []*client.NewsAPIArticle, sourceID uint) ([]*ExtractedArticle, error) {
	var extracted []*ExtractedArticle

	for _, apiArticle := range apiArticles {
		article, err := s.extractNewsAPIArticle(ctx, apiArticle, sourceID)
		if err != nil {
			logrus.WithError(err).Warnf("Failed to extract article: %s", apiArticle.Title)
			continue
		}

		// Check for duplicates
		if s.isDuplicate(ctx, article) {
			logrus.Debugf("Duplicate detected: %s", article.Title)
			continue
		}

		extracted = append(extracted, article)
	}

	return extracted, nil
}

// ExtractFromRSS extracts and processes articles from RSS feeds
func (s *DataExtractionService) ExtractFromRSS(ctx context.Context, rssItems []*client.RSSItem, sourceID uint) ([]*ExtractedArticle, error) {
	var extracted []*ExtractedArticle

	for _, item := range rssItems {
		article, err := s.extractRSSArticle(ctx, item, sourceID)
		if err != nil {
			logrus.WithError(err).Warnf("Failed to extract RSS item: %s", item.Title)
			continue
		}

		// Check for duplicates
		if s.isDuplicate(ctx, article) {
			logrus.Debugf("Duplicate detected: %s", article.Title)
			continue
		}

		extracted = append(extracted, article)
	}

	return extracted, nil
}

// extractNewsAPIArticle processes a NewsAPI article
func (s *DataExtractionService) extractNewsAPIArticle(ctx context.Context, apiArticle *client.NewsAPIArticle, sourceID uint) (*ExtractedArticle, error) {
	// Clean and normalize content
	title := s.cleanText(apiArticle.Title)
	content := s.cleanText(apiArticle.Content)
	if content == "" {
		content = s.cleanText(apiArticle.Description)
	}

	// Validate minimum content length
	if len(content) < 50 {
		return nil, fmt.Errorf("content too short")
	}

	// Extract full text for analysis
	fullText := title + " " + content

	// Extract financial symbols
	symbols := s.symbolExtractor.ExtractFromText(fullText)

	// Extract keywords
	keywords := s.keywordExtractor.Extract(fullText)

	// Classify category
	category := s.categoryClassifier.Classify(fullText)

	// Generate content hash
	contentHash := s.generateContentHash(title + content + apiArticle.URL)

	// Calculate relevance score
	relevanceScore := s.calculateRelevanceScore(fullText, symbols, keywords)

	// Generate summary (first 200 chars of content)
	summary := s.generateSummary(content)

	article := &ExtractedArticle{
		ID:               uuid.New(),
		SourceID:         sourceID,
		Title:            title,
		Content:          content,
		Summary:          summary,
		URL:              apiArticle.URL,
		ImageURL:         apiArticle.URLToImage,
		Author:           apiArticle.Author,
		Symbols:          symbols,
		Keywords:         keywords,
		PublishedAt:      apiArticle.PublishedAt,
		ContentHash:      contentHash,
		RelevanceScore:   relevanceScore,
		ProcessingStatus: string(model.StatusPending),
		Language:         s.detectLanguage(content),
		Category:         category,
	}

	return article, nil
}

// extractRSSArticle processes an RSS item
func (s *DataExtractionService) extractRSSArticle(ctx context.Context, item *client.RSSItem, sourceID uint) (*ExtractedArticle, error) {
	// Clean and normalize content
	title := s.cleanText(item.Title)
	content := s.cleanText(item.Description)

	// Validate minimum content length
	if len(content) < 50 {
		return nil, fmt.Errorf("content too short")
	}

	// Extract full text for analysis
	fullText := title + " " + content

	// Extract financial symbols
	symbols := s.symbolExtractor.ExtractFromText(fullText)

	// Extract keywords
	keywords := s.keywordExtractor.Extract(fullText)

	// Classify category
	category := s.categoryClassifier.Classify(fullText)

	// Generate content hash
	contentHash := s.generateContentHash(title + content + item.Link)

	// Calculate relevance score
	relevanceScore := s.calculateRelevanceScore(fullText, symbols, keywords)

	// Generate summary
	summary := s.generateSummary(content)

	article := &ExtractedArticle{
		ID:               uuid.New(),
		SourceID:         sourceID,
		Title:            title,
		Content:          content,
		Summary:          summary,
		URL:              item.Link,
		Symbols:          symbols,
		Keywords:         keywords,
		PublishedAt:      item.PubDate,
		ContentHash:      contentHash,
		RelevanceScore:   relevanceScore,
		ProcessingStatus: string(model.StatusPending),
		Language:         s.detectLanguage(content),
		Category:         category,
	}

	return article, nil
}

// SaveArticles saves extracted articles to database
func (s *DataExtractionService) SaveArticles(ctx context.Context, articles []*ExtractedArticle) (int, error) {
	savedCount := 0

	for _, article := range articles {
		dbArticle := &model.Article{
			ID:               article.ID,
			SourceID:         article.SourceID,
			Title:            article.Title,
			Content:          article.Content,
			URL:              article.URL,
			Symbols:          article.Symbols,
			PublishedAt:      article.PublishedAt,
			ProcessingStatus: model.StatusPending,
			RelevanceScore:   article.RelevanceScore,
			ContentHash:      article.ContentHash,
		}

		if err := s.articleRepo.Create(ctx, dbArticle); err != nil {
			logrus.WithError(err).Errorf("Failed to save article: %s", article.Title)
			continue
		}

		savedCount++
	}

	return savedCount, nil
}

// Helper methods

func (s *DataExtractionService) cleanText(text string) string {
	// Remove extra whitespace
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Remove special characters but keep basic punctuation
	text = regexp.MustCompile(`[^\w\s\.,!?;:()\-\'\"$%]`).ReplaceAllString(text, "")

	return text
}

func (s *DataExtractionService) generateContentHash(content string) string {
	hasher := sha256.New()
	hasher.Write([]byte(strings.ToLower(strings.TrimSpace(content))))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func (s *DataExtractionService) calculateRelevanceScore(text string, symbols []string, keywords []string) float64 {
	score := 0.0

	// Symbol presence
	if len(symbols) > 0 {
		score += 0.3
		if len(symbols) > 3 {
			score += 0.1
		}
	}

	// Keyword density
	if len(keywords) > 0 {
		keywordScore := float64(len(keywords)) * 0.05
		if keywordScore > 0.3 {
			keywordScore = 0.3
		}
		score += keywordScore
	}

	// Check for high-value keywords
	highValueKeywords := []string{"earnings", "merger", "acquisition", "ipo", "dividend", "bankruptcy"}
	textLower := strings.ToLower(text)
	for _, keyword := range highValueKeywords {
		if strings.Contains(textLower, keyword) {
			score += 0.1
		}
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (s *DataExtractionService) generateSummary(content string) string {
	if len(content) <= 200 {
		return content
	}

	// Take first 200 chars and find last sentence
	summary := content[:200]
	lastPeriod := strings.LastIndex(summary, ".")
	if lastPeriod > 50 {
		summary = summary[:lastPeriod+1]
	}

	return summary + "..."
}

func (s *DataExtractionService) detectLanguage(text string) string {
	// Simple language detection (can be enhanced with proper library)
	// For now, assume English
	return "en"
}

func (s *DataExtractionService) isDuplicate(ctx context.Context, article *ExtractedArticle) bool {
	modelArticle := &model.Article{
		Title:       article.Title,
		Content:     article.Content,
		ContentHash: article.ContentHash,
		PublishedAt: article.PublishedAt,
	}

	return s.deduplicationSvc.IsDuplicate(ctx, modelArticle)
}

// Helper constructors

func newSymbolExtractor() *SymbolExtractor {
	symbols := map[string]bool{
		"AAPL": true, "GOOGL": true, "MSFT": true, "AMZN": true, "TSLA": true,
		"META": true, "NVDA": true, "AMD": true, "NFLX": true, "CRM": true,
		"BTC": true, "ETH": true, "SPY": true, "QQQ": true, "GLD": true,
		"JPM": true, "BAC": true, "WFC": true, "GS": true, "C": true,
		"XOM": true, "CVX": true, "PG": true, "JNJ": true, "V": true,
		"DIS": true, "INTC": true, "CSCO": true, "IBM": true, "ORCL": true,
	}

	return &SymbolExtractor{financialSymbols: symbols}
}

func newKeywordExtractor() *KeywordExtractor {
	keywords := map[string]float64{
		"earnings":      1.0,
		"revenue":       0.9,
		"profit":        0.9,
		"loss":          0.8,
		"merger":        1.0,
		"acquisition":   1.0,
		"ipo":           1.0,
		"dividend":      0.8,
		"bankruptcy":    1.0,
		"layoffs":       0.7,
		"expansion":     0.7,
		"investment":    0.6,
		"stock":         0.5,
		"market":        0.5,
		"trading":       0.5,
		"nasdaq":        0.6,
		"dow":           0.6,
		"s&p":           0.6,
		"bull":          0.6,
		"bear":          0.6,
		"rally":         0.7,
		"crash":         0.9,
		"volatility":    0.7,
		"inflation":     0.8,
		"interest rate": 0.9,
		"fed":           0.8,
		"gdp":           0.7,
	}

	return &KeywordExtractor{financialKeywords: keywords}
}

func newCategoryClassifier() *CategoryClassifier {
	categories := map[string][]string{
		"earnings":           {"earnings", "revenue", "profit", "quarterly", "annual report"},
		"merger_acquisition": {"merger", "acquisition", "takeover", "buyout", "deal"},
		"ipo":                {"ipo", "initial public offering", "listing", "debut"},
		"market":             {"market", "dow", "nasdaq", "s&p", "index", "trading"},
		"regulatory":         {"sec", "regulation", "compliance", "lawsuit", "investigation"},
		"technology":         {"ai", "artificial intelligence", "blockchain", "crypto", "tech"},
		"energy":             {"oil", "gas", "energy", "renewable", "electric"},
		"healthcare":         {"pharma", "drug", "fda", "healthcare", "medical"},
		"finance":            {"bank", "lending", "credit", "interest", "mortgage"},
	}

	return &CategoryClassifier{categories: categories}
}

func (ke *KeywordExtractor) Extract(text string) []string {
	textLower := strings.ToLower(text)
	var extractedKeywords []string

	for keyword, score := range ke.financialKeywords {
		if strings.Contains(textLower, keyword) && score >= 0.6 {
			extractedKeywords = append(extractedKeywords, keyword)
		}
	}

	// Limit to top 10 keywords
	if len(extractedKeywords) > 10 {
		extractedKeywords = extractedKeywords[:10]
	}

	return extractedKeywords
}

func (cc *CategoryClassifier) Classify(text string) string {
	textLower := strings.ToLower(text)
	maxScore := 0.0
	bestCategory := "general"

	for category, keywords := range cc.categories {
		score := 0.0
		for _, keyword := range keywords {
			if strings.Contains(textLower, keyword) {
				score += 1.0
			}
		}

		if score > maxScore {
			maxScore = score
			bestCategory = category
		}
	}

	return bestCategory
}
