// services/news-ingestion/internal/service/deduplication_service.go
package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strings"
	"unicode"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/repository"
)

type DeduplicationService interface {
	IsDuplicate(ctx context.Context, article *model.Article) bool
	GenerateContentHash(content string) string
	CalculateSimilarity(content1, content2 string) float64
	FindSimilarArticles(ctx context.Context, article *model.Article, threshold float64) ([]*model.Article, error)
}

type deduplicationService struct {
	articleRepo repository.ArticleRepository
}

func NewDeduplicationService(articleRepo repository.ArticleRepository) DeduplicationService {
	return &deduplicationService{
		articleRepo: articleRepo,
	}
}

func (s *deduplicationService) IsDuplicate(ctx context.Context, article *model.Article) bool {
	// Check if content hash already exists
	if article.ContentHash == "" {
		article.ContentHash = s.GenerateContentHash(article.Title + article.Content)
	}

	existing, err := s.articleRepo.GetByContentHash(ctx, article.ContentHash)
	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.WithError(err).Error("Failed to check for duplicate by content hash")
		return false
	}

	if existing != nil {
		logrus.Debugf("Exact duplicate found for article: %s", article.Title)
		return true
	}

	// Check for near-duplicates using similarity scoring
	similar, err := s.FindSimilarArticles(ctx, article, 0.85) // 85% similarity threshold
	if err != nil {
		logrus.WithError(err).Error("Failed to check for similar articles")
		return false
	}

	if len(similar) > 0 {
		logrus.Debugf("Similar article found for: %s", article.Title)
		return true
	}

	return false
}

func (s *deduplicationService) GenerateContentHash(content string) string {
	// Normalize content for hashing
	normalized := s.normalizeContent(content)

	hasher := md5.New()
	hasher.Write([]byte(normalized))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (s *deduplicationService) CalculateSimilarity(content1, content2 string) float64 {
	// Normalize both contents
	norm1 := s.normalizeContent(content1)
	norm2 := s.normalizeContent(content2)

	// Use simple word-based Jaccard similarity
	words1 := s.extractWords(norm1)
	words2 := s.extractWords(norm2)

	// Convert to sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, word := range words1 {
		set1[word] = true
	}

	for _, word := range words2 {
		set2[word] = true
	}

	// Calculate intersection and union
	intersection := 0
	union := make(map[string]bool)

	for word := range set1 {
		union[word] = true
		if set2[word] {
			intersection++
		}
	}

	for word := range set2 {
		union[word] = true
	}

	if len(union) == 0 {
		return 0.0
	}

	return float64(intersection) / float64(len(union))
}

func (s *deduplicationService) FindSimilarArticles(ctx context.Context, article *model.Article, threshold float64) ([]*model.Article, error) {
	// Get recent articles to compare against (last 7 days)
	endTime := article.PublishedAt
	startTime := endTime.AddDate(0, 0, -7)

	candidates, err := s.articleRepo.GetByDateRange(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var similarArticles []*model.Article
	targetContent := article.Title + " " + article.Content

	for _, candidate := range candidates {
		if candidate.ID == article.ID {
			continue // Skip self
		}

		candidateContent := candidate.Title + " " + candidate.Content
		similarity := s.CalculateSimilarity(targetContent, candidateContent)

		if similarity >= threshold {
			similarArticles = append(similarArticles, candidate)
		}
	}

	return similarArticles, nil
}

func (s *deduplicationService) normalizeContent(content string) string {
	// Convert to lowercase
	normalized := strings.ToLower(content)

	// Remove extra whitespace and special characters
	var builder strings.Builder
	var prevSpace bool

	for _, r := range normalized {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			prevSpace = false
		} else if unicode.IsSpace(r) && !prevSpace {
			builder.WriteRune(' ')
			prevSpace = true
		}
	}

	return strings.TrimSpace(builder.String())
}

func (s *deduplicationService) extractWords(content string) []string {
	words := strings.Fields(content)

	// Filter out common stop words and short words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
	}

	var filteredWords []string
	for _, word := range words {
		if len(word) >= 3 && !stopWords[word] {
			filteredWords = append(filteredWords, word)
		}
	}

	return filteredWords
}
