// services/nlp-processing/internal/service/topic_service.go
package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
)

type TopicService interface {
	ClassifyTopic(ctx context.Context, articleID, text string) (*model.TopicClassification, error)
	BatchClassifyTopics(ctx context.Context, articles []model.AnalysisRequest) (map[string]*model.TopicClassification, error)
	LoadModel(ctx context.Context) error
	IsModelLoaded() bool
}

type topicService struct {
	modelLoaded bool
	modelPath   string
	categories  []string

	// Placeholder for actual topic classification model
	// In production, this would be replaced with actual model interface
	// topicModel *topic.Classifier
}

type TopicConfig struct {
	ModelPath  string   `mapstructure:"model_path"`
	Categories []string `mapstructure:"categories"`
}

type TopicScore struct {
	Topic      string
	Confidence float32
}

func NewTopicService(config TopicConfig) TopicService {
	return &topicService{
		modelLoaded: false,
		modelPath:   config.ModelPath,
		categories:  config.Categories,
	}
}

func (t *topicService) LoadModel(ctx context.Context) error {
	logrus.Info("Loading topic classification model...")

	// TODO: Replace with actual model loading when models are provided
	// Example implementation structure:
	/*
		classifier, err := topic.LoadClassifier(t.modelPath)
		if err != nil {
			return fmt.Errorf("failed to load topic classifier: %w", err)
		}
		t.topicModel = classifier
	*/

	// Simulate model loading time
	time.Sleep(1 * time.Second)

	t.modelLoaded = true
	logrus.Info("Topic classification model loaded successfully")

	return nil
}

func (t *topicService) IsModelLoaded() bool {
	return t.modelLoaded
}

func (t *topicService) ClassifyTopic(ctx context.Context, articleID, text string) (*model.TopicClassification, error) {
	if !t.modelLoaded {
		return nil, fmt.Errorf("topic classification model not loaded")
	}

	// TODO: Replace with actual model inference when models are provided
	// Example implementation structure:
	/*
		predictions, err := t.topicModel.Predict(text)
		if err != nil {
			return nil, fmt.Errorf("topic classification failed: %w", err)
		}
	*/

	// Stub implementation - classify topics using keyword matching
	topicScores := t.stubTopicClassification(text)

	// Sort by confidence and get primary topic
	sort.Slice(topicScores, func(i, j int) bool {
		return topicScores[i].Confidence > topicScores[j].Confidence
	})

	primaryTopic := "General"
	primaryConfidence := float32(0.5)
	if len(topicScores) > 0 {
		primaryTopic = topicScores[0].Topic
		primaryConfidence = topicScores[0].Confidence
	}

	// Extract keywords
	keywords := t.extractKeywords(text)

	// Calculate urgency score
	urgencyScore := t.calculateUrgencyScore(text, primaryTopic)

	// Determine if breaking news
	breakingNews := t.isBreakingNews(text, urgencyScore)

	classification := &model.TopicClassification{
		ArticleID:              articleID,
		PrimaryTopic:           primaryTopic,
		PrimaryTopicConfidence: primaryConfidence,
		Keywords:               keywords,
		UrgencyScore:           urgencyScore,
		BreakingNewsIndicator:  breakingNews,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	logrus.WithFields(logrus.Fields{
		"article_id":    articleID,
		"primary_topic": primaryTopic,
		"confidence":    primaryConfidence,
		"urgency":       urgencyScore,
		"breaking_news": breakingNews,
	}).Debug("Topic classification completed")

	return classification, nil
}

func (t *topicService) BatchClassifyTopics(ctx context.Context, articles []model.AnalysisRequest) (map[string]*model.TopicClassification, error) {
	if !t.modelLoaded {
		return nil, fmt.Errorf("topic classification model not loaded")
	}

	results := make(map[string]*model.TopicClassification)

	for _, article := range articles {
		classification, err := t.ClassifyTopic(ctx, article.ArticleID, article.Title+" "+article.Content)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to classify topic for article %s", article.ArticleID)
			continue
		}
		results[article.ArticleID] = classification
	}

	return results, nil
}

// Stub implementation for topic classification until actual models are provided
func (t *topicService) stubTopicClassification(text string) []TopicScore {
	text = strings.ToLower(text)
	scores := make([]TopicScore, 0)

	// Define topic keywords
	topicKeywords := map[string][]string{
		"Earnings":               {"earnings", "revenue", "profit", "eps", "quarterly", "q1", "q2", "q3", "q4", "financial results"},
		"Mergers & Acquisitions": {"merger", "acquisition", "acquire", "bought", "deal", "takeover", "buyout"},
		"Market Analysis":        {"market", "trading", "stocks", "dow", "nasdaq", "s&p", "index", "bull", "bear"},
		"Economic Indicators":    {"gdp", "inflation", "fed", "interest rate", "unemployment", "cpi", "economic"},
		"Regulatory News":        {"sec", "fda", "regulation", "compliance", "legal", "court", "lawsuit", "fine"},
		"Company News":           {"ceo", "executive", "management", "strategy", "product", "launch", "announcement"},
		"Crypto":                 {"bitcoin", "cryptocurrency", "crypto", "blockchain", "eth", "ethereum", "digital currency"},
		"Commodities":            {"oil", "gold", "silver", "commodity", "energy", "natural gas", "crude"},
	}

	// Calculate scores for each topic
	for topic, keywords := range topicKeywords {
		matchCount := 0
		totalKeywords := len(keywords)

		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				matchCount++
			}
		}

		if matchCount > 0 {
			confidence := float32(matchCount) / float32(totalKeywords)
			if confidence > 0.1 { // Minimum threshold
				scores = append(scores, TopicScore{
					Topic:      topic,
					Confidence: confidence,
				})
			}
		}
	}

	return scores
}

func (t *topicService) extractKeywords(text string) []string {
	// Simple keyword extraction based on frequency and financial relevance
	words := strings.Fields(strings.ToLower(text))
	wordCount := make(map[string]int)

	// Count word frequencies
	for _, word := range words {
		// Clean word
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 3 { // Only consider words longer than 3 characters
			wordCount[word]++
		}
	}

	// Define financial keywords to prioritize
	financialKeywords := map[string]bool{
		"earnings": true, "revenue": true, "profit": true, "loss": true,
		"stock": true, "market": true, "trading": true, "investment": true,
		"financial": true, "economic": true, "quarterly": true, "annual": true,
		"merger": true, "acquisition": true, "ceo": true, "executive": true,
		"regulation": true, "compliance": true, "bitcoin": true, "crypto": true,
	}

	// Extract top keywords
	type WordScore struct {
		Word  string
		Score int
	}

	var wordScores []WordScore
	for word, count := range wordCount {
		score := count
		if financialKeywords[word] {
			score *= 2 // Boost financial keywords
		}
		wordScores = append(wordScores, WordScore{Word: word, Score: score})
	}

	// Sort by score and take top keywords
	sort.Slice(wordScores, func(i, j int) bool {
		return wordScores[i].Score > wordScores[j].Score
	})

	keywords := make([]string, 0, 10)
	for i, ws := range wordScores {
		if i >= 10 || ws.Score < 2 { // Limit to top 10 keywords with minimum frequency
			break
		}
		keywords = append(keywords, ws.Word)
	}

	return keywords
}

func (t *topicService) calculateUrgencyScore(text, primaryTopic string) float32 {
	text = strings.ToLower(text)
	urgencyScore := float32(0.0)

	// Urgent keywords and phrases
	urgentKeywords := []string{
		"breaking", "urgent", "alert", "emergency", "crash", "plummet",
		"surge", "soar", "halt", "suspend", "investigate", "probe",
		"lawsuit", "scandal", "bankruptcy", "default", "crisis",
	}

	for _, keyword := range urgentKeywords {
		if strings.Contains(text, keyword) {
			urgencyScore += 0.2
		}
	}

	// Time-sensitive indicators
	timeKeywords := []string{
		"today", "now", "just", "immediate", "sudden", "unexpected",
		"minutes ago", "hours ago", "this morning", "this afternoon",
	}

	for _, keyword := range timeKeywords {
		if strings.Contains(text, keyword) {
			urgencyScore += 0.1
		}
	}

	// Topic-specific urgency boosts
	switch primaryTopic {
	case "Regulatory News":
		urgencyScore += 0.2
	case "Mergers & Acquisitions":
		urgencyScore += 0.15
	case "Market Analysis":
		if strings.Contains(text, "crash") || strings.Contains(text, "surge") {
			urgencyScore += 0.3
		}
	}

	// Cap at 1.0
	if urgencyScore > 1.0 {
		urgencyScore = 1.0
	}

	return urgencyScore
}

func (t *topicService) isBreakingNews(text string, urgencyScore float32) uint8 {
	if urgencyScore > 0.7 {
		return 1
	}

	text = strings.ToLower(text)
	breakingPhrases := []string{
		"breaking news", "breaking:", "urgent:", "alert:",
		"just in:", "developing:", "exclusive:",
	}

	for _, phrase := range breakingPhrases {
		if strings.Contains(text, phrase) {
			return 1
		}
	}

	return 0
}
