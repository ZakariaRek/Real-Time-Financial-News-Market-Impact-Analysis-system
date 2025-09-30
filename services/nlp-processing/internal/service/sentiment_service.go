// services/nlp-processing/internal/service/sentiment_service.go
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
	"github.com/sirupsen/logrus"
)

type SentimentService interface {
	AnalyzeSentiment(ctx context.Context, articleID, text string, symbols []string) (*model.SentimentAnalysis, error)
	BatchAnalyzeSentiment(ctx context.Context, articles []model.AnalysisRequest) ([]*model.SentimentAnalysis, error)
	LoadModel(ctx context.Context) error
	IsModelLoaded() bool
}

type sentimentService struct {
	modelLoaded  bool
	modelVersion string
}

func NewSentimentService() SentimentService {
	return &sentimentService{
		modelLoaded:  false,
		modelVersion: "textblob-fallback-1.0",
	}
}

func (s *sentimentService) LoadModel(ctx context.Context) error {
	logrus.Info("Initializing sentiment analysis (TextBlob-based)")
	time.Sleep(1 * time.Second)
	s.modelLoaded = true
	logrus.Info("Sentiment analysis ready")
	return nil
}

func (s *sentimentService) IsModelLoaded() bool {
	return s.modelLoaded
}

func (s *sentimentService) AnalyzeSentiment(ctx context.Context, articleID, text string, symbols []string) (*model.SentimentAnalysis, error) {
	if !s.modelLoaded {
		return nil, fmt.Errorf("sentiment model not loaded")
	}

	// Calculate sentiment score (-1 to 1)
	compoundScore, confidence := s.calculateSentiment(text)

	// Determine primary S&P 500 symbol
	primarySymbol := s.extractPrimarySymbol(symbols, text)

	analysis := &model.SentimentAnalysis{
		ArticleID:         articleID,
		AnalysisTimestamp: time.Now().UTC(),
		CompoundScore:     compoundScore,
		Confidence:        confidence,
		PrimarySymbol:     primarySymbol,
		ModelVersion:      s.modelVersion,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	logrus.WithFields(logrus.Fields{
		"article_id": articleID,
		"sentiment":  compoundScore,
		"confidence": confidence,
		"symbol":     primarySymbol,
	}).Debug("Sentiment analysis completed")

	return analysis, nil
}

func (s *sentimentService) BatchAnalyzeSentiment(ctx context.Context, articles []model.AnalysisRequest) ([]*model.SentimentAnalysis, error) {
	results := make([]*model.SentimentAnalysis, 0, len(articles))

	for _, article := range articles {
		analysis, err := s.AnalyzeSentiment(ctx, article.ArticleID, article.Title+" "+article.Content, article.Symbols)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to analyze sentiment for article %s", article.ArticleID)
			continue
		}
		results = append(results, analysis)
	}

	return results, nil
}

// calculateSentiment uses keyword-based sentiment analysis (similar to TextBlob)
func (s *sentimentService) calculateSentiment(text string) (float32, float32) {
	text = strings.ToLower(text)

	// S&P 500 specific positive keywords
	positiveWords := map[string]float32{
		"gain": 0.5, "gains": 0.5, "up": 0.3, "rise": 0.4, "rises": 0.4, "rally": 0.6,
		"surge": 0.7, "soar": 0.7, "profit": 0.6, "earnings": 0.4, "beat": 0.5,
		"strong": 0.5, "growth": 0.5, "positive": 0.4, "bullish": 0.7, "bull": 0.6,
		"high": 0.3, "record": 0.5, "outperform": 0.6, "upgrade": 0.5, "buy": 0.4,
	}

	// S&P 500 specific negative keywords
	negativeWords := map[string]float32{
		"loss": -0.6, "losses": -0.6, "down": -0.3, "fall": -0.4, "falls": -0.4,
		"decline": -0.5, "drop": -0.5, "plunge": -0.7, "crash": -0.8, "bearish": -0.7,
		"bear": -0.6, "weak": -0.5, "negative": -0.4, "miss": -0.5, "disappoint": -0.6,
		"concern": -0.4, "risk": -0.3, "volatility": -0.4, "downgrade": -0.5, "sell": -0.4,
	}

	words := strings.Fields(text)
	var totalScore float32
	var matches int

	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")

		if score, exists := positiveWords[word]; exists {
			totalScore += score
			matches++
		} else if score, exists := negativeWords[word]; exists {
			totalScore += score
			matches++
		}
	}

	// Calculate compound score
	var compoundScore float32
	if matches > 0 {
		compoundScore = totalScore / float32(matches)
	} else {
		compoundScore = 0.0
	}

	// Normalize between -1 and 1
	if compoundScore > 1.0 {
		compoundScore = 1.0
	} else if compoundScore < -1.0 {
		compoundScore = -1.0
	}

	// Calculate confidence based on number of matches
	confidence := float32(matches) / float32(len(words))
	if confidence > 0.9 {
		confidence = 0.9
	}
	if confidence < 0.3 {
		confidence = 0.3
	}

	return compoundScore, confidence
}

// extractPrimarySymbol identifies the most relevant S&P 500 symbol
func (s *sentimentService) extractPrimarySymbol(symbols []string, text string) string {
	if len(symbols) == 0 {
		return "SPY" // Default to S&P 500 ETF
	}

	// S&P 500 major components priority
	sp500Priority := map[string]int{
		"SPY": 100, "SPX": 100, // S&P 500 itself
		"AAPL": 10, "MSFT": 10, "GOOGL": 10, "AMZN": 10, "NVDA": 10,
		"TSLA": 9, "META": 9, "BRK.B": 9, "UNH": 8, "JPM": 8,
	}

	textLower := strings.ToLower(text)
	bestSymbol := symbols[0]
	bestScore := 0

	for _, symbol := range symbols {
		score := sp500Priority[symbol]

		// Boost score if symbol is mentioned in text
		if strings.Contains(textLower, strings.ToLower(symbol)) {
			score += 5
		}

		if score > bestScore {
			bestScore = score
			bestSymbol = symbol
		}
	}

	return bestSymbol
}
