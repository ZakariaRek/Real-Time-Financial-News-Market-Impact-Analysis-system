// services/nlp-processing/internal/service/sentiment_service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
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
	classifier   *NaiveBayesClassifier
	modelLoaded  bool
	modelVersion string
}

// NaiveBayesClassifier implements Naive Bayes for sentiment analysis
type NaiveBayesClassifier struct {
	// Word counts per class
	PositiveWords map[string]int
	NeutralWords  map[string]int
	NegativeWords map[string]int

	// Document counts per class
	PositiveDocs int
	NeutralDocs  int
	NegativeDocs int

	// Total vocabulary size
	VocabSize int

	// Priors (log probabilities)
	PositivePrior float64
	NeutralPrior  float64
	NegativePrior float64
}

// ModelData for saving/loading
type ModelData struct {
	PositiveWords map[string]int `json:"positive_words"`
	NeutralWords  map[string]int `json:"neutral_words"`
	NegativeWords map[string]int `json:"negative_words"`
	PositiveDocs  int            `json:"positive_docs"`
	NeutralDocs   int            `json:"neutral_docs"`
	NegativeDocs  int            `json:"negative_docs"`
	VocabSize     int            `json:"vocab_size"`
}

func NewSentimentService() SentimentService {
	return &sentimentService{
		modelLoaded:  false,
		modelVersion: "naive-bayes-financial-1.0",
	}
}

func (s *sentimentService) LoadModel(ctx context.Context) error {
	logrus.Info("Loading Pure Go sentiment model...")

	// Try to load pre-trained model
	modelPath := "./models/sentiment/naive_bayes.json"
	classifier, err := LoadNaiveBayesModel(modelPath)
	if err != nil {
		// If no pre-trained model, create and train a new one
		logrus.Info("No pre-trained model found, training new model with financial data...")
		classifier = TrainDefaultFinancialSentimentModel()

		// Save the model
		os.MkdirAll("./models/sentiment", 0755)
		if err := SaveNaiveBayesModel(classifier, modelPath); err != nil {
			logrus.WithError(err).Warn("Failed to save trained model")
		} else {
			logrus.Infof("Model saved to %s", modelPath)
		}
	}

	s.classifier = classifier
	s.modelLoaded = true
	logrus.Info("Pure Go sentiment model loaded successfully")
	logrus.Infof("Model trained on %d documents (%d positive, %d neutral, %d negative)",
		classifier.PositiveDocs+classifier.NeutralDocs+classifier.NegativeDocs,
		classifier.PositiveDocs, classifier.NeutralDocs, classifier.NegativeDocs)
	return nil
}

func (s *sentimentService) IsModelLoaded() bool {
	return s.modelLoaded
}

func (s *sentimentService) AnalyzeSentiment(ctx context.Context, articleID, text string, symbols []string) (*model.SentimentAnalysis, error) {
	if !s.modelLoaded {
		return nil, fmt.Errorf("sentiment model not loaded")
	}

	// Predict sentiment
	sentiment, confidence := s.classifier.Predict(text)

	// Convert sentiment to compound score (-1 to 1)
	var compoundScore float32
	switch sentiment {
	case "positive":
		compoundScore = confidence
	case "negative":
		compoundScore = -confidence
	case "neutral":
		compoundScore = 0.0
	}

	// Extract primary symbol
	primarySymbol := extractPrimarySymbol(symbols, text)

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
		"sentiment":  sentiment,
		"score":      compoundScore,
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

// NewNaiveBayesClassifier creates a new classifier
func NewNaiveBayesClassifier() *NaiveBayesClassifier {
	return &NaiveBayesClassifier{
		PositiveWords: make(map[string]int),
		NeutralWords:  make(map[string]int),
		NegativeWords: make(map[string]int),
	}
}

// Train trains the classifier on labeled data
func (nb *NaiveBayesClassifier) Train(texts []string, labels []string) {
	// Count documents per class
	for _, label := range labels {
		switch label {
		case "positive":
			nb.PositiveDocs++
		case "neutral":
			nb.NeutralDocs++
		case "negative":
			nb.NegativeDocs++
		}
	}

	// Count words per class
	vocabSet := make(map[string]bool)
	for i, text := range texts {
		words := tokenize(text)
		label := labels[i]

		for _, word := range words {
			vocabSet[word] = true

			switch label {
			case "positive":
				nb.PositiveWords[word]++
			case "neutral":
				nb.NeutralWords[word]++
			case "negative":
				nb.NegativeWords[word]++
			}
		}
	}

	nb.VocabSize = len(vocabSet)

	// Calculate priors (log probabilities)
	totalDocs := float64(nb.PositiveDocs + nb.NeutralDocs + nb.NegativeDocs)
	nb.PositivePrior = math.Log(float64(nb.PositiveDocs) / totalDocs)
	nb.NeutralPrior = math.Log(float64(nb.NeutralDocs) / totalDocs)
	nb.NegativePrior = math.Log(float64(nb.NegativeDocs) / totalDocs)
}

// Predict predicts sentiment for a text
func (nb *NaiveBayesClassifier) Predict(text string) (string, float32) {
	words := tokenize(text)

	// Calculate log probabilities for each class
	positiveScore := nb.PositivePrior
	neutralScore := nb.NeutralPrior
	negativeScore := nb.NegativePrior

	// Total words per class (for Laplace smoothing)
	totalPositive := 0
	totalNeutral := 0
	totalNegative := 0

	for _, count := range nb.PositiveWords {
		totalPositive += count
	}
	for _, count := range nb.NeutralWords {
		totalNeutral += count
	}
	for _, count := range nb.NegativeWords {
		totalNegative += count
	}

	// Calculate likelihood for each word
	for _, word := range words {
		// Laplace smoothing: add 1 to all counts
		posProb := float64(nb.PositiveWords[word]+1) / float64(totalPositive+nb.VocabSize)
		neuProb := float64(nb.NeutralWords[word]+1) / float64(totalNeutral+nb.VocabSize)
		negProb := float64(nb.NegativeWords[word]+1) / float64(totalNegative+nb.VocabSize)

		positiveScore += math.Log(posProb)
		neutralScore += math.Log(neuProb)
		negativeScore += math.Log(negProb)
	}

	// Convert log probabilities to probabilities
	scores := []float64{positiveScore, neutralScore, negativeScore}
	probs := softmaxFloat64(scores)

	// Find max probability
	maxProb := probs[0]
	maxIndex := 0
	for i, prob := range probs[1:] {
		if prob > maxProb {
			maxProb = prob
			maxIndex = i + 1
		}
	}

	labels := []string{"positive", "neutral", "negative"}
	return labels[maxIndex], float32(maxProb)
}

// TrainDefaultFinancialSentimentModel creates a model with financial training data
func TrainDefaultFinancialSentimentModel() *NaiveBayesClassifier {
	// Comprehensive training data: financial news headlines and their sentiment
	texts := []string{
		// Positive examples (60)
		"stock surges to record high on strong earnings beat",
		"company reports revenue growth exceeding analyst expectations",
		"shares rally after positive guidance for next quarter",
		"merger deal announced at premium valuation",
		"dividend increased by 15 percent showing confidence",
		"profit margins expand as costs decline",
		"bullish outlook drives investor enthusiasm",
		"stock price soars on breakthrough innovation",
		"quarterly results exceed all projections",
		"market cap reaches new milestone",
		"outstanding performance drives gains",
		"revenue jump signals strong demand",
		"analysts upgrade rating to strong buy",
		"stock outperforms market benchmarks significantly",
		"strong fundamentals support upward trend",
		"record profits announced in latest quarter",
		"share buyback program boosts investor returns",
		"expansion into new markets drives growth",
		"breakthrough product launch exceeds expectations",
		"strategic partnership deal creates significant value",
		"cost cutting measures improve profitability dramatically",
		"market share gains drive revenue growth",
		"innovative technology leads sector performance",
		"strong cash flow generation continues",
		"debt reduction improves balance sheet strength",
		"new CEO brings fresh optimism to company",
		"acquisition enhances competitive position",
		"subscriber growth accelerates beyond forecasts",
		"pricing power demonstrated in latest results",
		"operational efficiency drives margins higher",
		"customer satisfaction scores improve markedly",
		"brand value strengthens market position",
		"supply chain optimization yields major savings",
		"digital transformation shows early success",
		"international expansion gains strong momentum",
		"product portfolio strengthening significantly",
		"talent acquisition strengthens leadership team",
		"sustainability initiatives attract new investors",
		"market leadership position solidifies further",
		"innovation pipeline looks very promising",
		"strategic vision resonates with investors",
		"execution excellence drives outstanding results",
		"margin improvement trend continues strongly",
		"revenue diversification reduces business risk",
		"customer retention rates improve substantially",
		"competitive advantages widen significantly",
		"scalability demonstrated in rapid growth",
		"unit economics improve dramatically",
		"network effects strengthen economic moat",
		"recurring revenue model performs exceptionally well",
		"earnings growth accelerates impressively",
		"stock rebounds strongly from recent lows",
		"investor confidence returns after positive news",
		"company beats expectations across all metrics",
		"management guidance raises optimism",
		"market sentiment turns decidedly positive",
		"trading volume surges on good news",
		"institutional investors increase holdings",
		"short interest declines significantly",
		"technical indicators signal strong momentum",

		// Negative examples (60)
		"stock plunges on disappointing earnings report",
		"company warns of significantly lower guidance amid challenges",
		"shares fall sharply after regulatory concerns emerge",
		"losses widen dramatically as revenue misses estimates",
		"major lawsuit threatens to impact profitability",
		"stock crashes on mounting bankruptcy fears",
		"management cuts forecast citing strong headwinds",
		"declining margins pressure earnings severely",
		"competition erodes market share rapidly",
		"debt levels raise serious solvency concerns",
		"downgrade to sell rating triggers massive selloff",
		"weak demand signals serious trouble ahead",
		"investigation launched into accounting practices",
		"massive layoffs announced amid restructuring",
		"revenue decline accelerates in latest quarter",
		"cash burn rate raises going concern questions",
		"product recall damages brand reputation severely",
		"cyber security breach exposes major vulnerabilities",
		"key executive departure creates uncertainty",
		"major contract loss impacts future revenue",
		"regulatory fine significantly impacts earnings",
		"failed product launch disappoints investors badly",
		"market share losses accelerate dramatically",
		"pricing pressure compresses margins severely",
		"inventory buildup signals weak demand",
		"customer churn rates increase alarmingly",
		"competitive threats intensify significantly",
		"technology obsolescence risk emerges",
		"supply chain disruptions impact production badly",
		"quality issues lead to customer complaints",
		"strategic missteps hurt performance badly",
		"integration challenges follow troubled acquisition",
		"credit rating downgrade announced",
		"pension obligations weigh heavily on finances",
		"environmental violations trigger major fines",
		"unionization efforts create significant uncertainty",
		"patent expiration threatens revenue stream",
		"restatement of financial results required",
		"going private transaction at steep discount",
		"asset impairment charges announced",
		"goodwill writedown impacts balance sheet",
		"restructuring costs exceed initial expectations",
		"market saturation severely limits growth",
		"subscriber losses accelerate rapidly",
		"advertising revenue declines sharply",
		"capital expenditure requirements increase dramatically",
		"working capital needs strain finances",
		"contingent liabilities surface unexpectedly",
		"covenant violations trigger serious concerns",
		"liquidity position deteriorates rapidly",
		"stock hits new 52-week low",
		"market confidence evaporates quickly",
		"analysts slash price targets dramatically",
		"insider selling raises red flags",
		"class action lawsuit filed against company",
		"regulatory probe expands significantly",
		"debt covenant breach announced",
		"management credibility questioned seriously",
		"earnings warning shocks investors",
		"guidance withdrawal spooks market",

		// Neutral examples (60)
		"company maintains quarterly dividend unchanged",
		"federal reserve holds interest rates steady as expected",
		"trading volume remains in line with average",
		"management provides routine business update",
		"stock price unchanged in afternoon trading session",
		"analyst maintains hold rating on shares",
		"company announces routine executive change",
		"market closes mixed with no clear direction",
		"earnings exactly in line with consensus estimates",
		"business conditions remain generally stable",
		"company files standard regulatory report",
		"trading continues within normal historical range",
		"management discusses ongoing strategic initiatives",
		"stock follows broader market trend closely",
		"quarter ends without any major surprises",
		"guidance reaffirmed for current fiscal year",
		"routine board meeting held as scheduled",
		"annual shareholder meeting scheduled",
		"standard product update released on schedule",
		"regular maintenance period scheduled",
		"normal course issuer bid announced",
		"typical seasonal patterns observed",
		"quarterly conference call scheduled",
		"investor day event planned for next month",
		"ordinary course of business activities continue",
		"standard compliance filing submitted",
		"routine external audit completed",
		"regular dividend payment date set",
		"fiscal year end approaching as planned",
		"annual report published on time",
		"proxy statement filed with regulators",
		"insider transactions reported routinely",
		"institutional ownership remains unchanged",
		"analyst coverage initiated at neutral rating",
		"peer comparison shows similar performance",
		"industry trends continue as expected",
		"macro economic environment remains stable",
		"sector performs in line with expectations",
		"valuation metrics in normal historical range",
		"beta coefficient remains stable",
		"correlation with market indices maintained",
		"volatility within expected normal range",
		"trading patterns remain typical",
		"volume profile within normal parameters",
		"bid ask spread unchanged from average",
		"market depth remains adequate",
		"liquidity conditions satisfactory",
		"share float unchanged from previous",
		"shares outstanding remain stable",
		"capital structure maintained as planned",
		"debt maturity schedule on track",
		"credit metrics remain stable",
		"cash position adequate for operations",
		"working capital ratios normal",
		"standard operating procedures followed",
		"routine compliance activities ongoing",
		"normal business cycle patterns observed",
		"market conditions stable overall",
		"industry dynamics unchanged",
		"competitive landscape stable",
	}

	labels := make([]string, len(texts))
	// Positive labels (60)
	for i := 0; i < 60; i++ {
		labels[i] = "positive"
	}
	// Negative labels (60)
	for i := 60; i < 120; i++ {
		labels[i] = "negative"
	}
	// Neutral labels (60)
	for i := 120; i < 180; i++ {
		labels[i] = "neutral"
	}

	classifier := NewNaiveBayesClassifier()
	classifier.Train(texts, labels)

	logrus.Info("Trained sentiment model with 180 financial examples")
	return classifier
}

// Helper functions

func tokenize(text string) []string {
	// Convert to lowercase and split on whitespace/punctuation
	text = strings.ToLower(text)

	var tokens []string
	var currentToken strings.Builder

	for _, char := range text {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			currentToken.WriteRune(char)
		} else {
			if currentToken.Len() > 0 {
				token := currentToken.String()
				// Filter out very short words and common stop words
				if len(token) > 2 && !isStopWord(token) {
					tokens = append(tokens, token)
				}
				currentToken.Reset()
			}
		}
	}

	if currentToken.Len() > 0 {
		token := currentToken.String()
		if len(token) > 2 && !isStopWord(token) {
			tokens = append(tokens, token)
		}
	}

	return tokens
}

func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "and": true, "for": true, "are": true,
		"but": true, "not": true, "you": true, "all": true,
		"can": true, "has": true, "had": true, "was": true,
		"with": true, "that": true, "this": true, "from": true,
		"will": true, "have": true, "been": true, "were": true,
	}
	return stopWords[word]
}

func softmaxFloat64(values []float64) []float64 {
	// Find max for numerical stability
	maxVal := values[0]
	for _, v := range values[1:] {
		if v > maxVal {
			maxVal = v
		}
	}

	// Compute exp and sum
	var expSum float64
	expValues := make([]float64, len(values))
	for i, v := range values {
		expValues[i] = math.Exp(v - maxVal)
		expSum += expValues[i]
	}

	// Normalize
	for i := range expValues {
		expValues[i] /= expSum
	}

	return expValues
}

func extractPrimarySymbol(symbols []string, text string) string {
	if len(symbols) == 0 {
		return "SPY"
	}

	sp500Priority := map[string]int{
		"SPY": 100, "SPX": 100,
		"AAPL": 10, "MSFT": 10, "GOOGL": 10, "AMZN": 10, "NVDA": 10,
		"TSLA": 9, "META": 9, "BRK.B": 9, "UNH": 8, "JPM": 8,
	}

	textLower := strings.ToLower(text)
	bestSymbol := symbols[0]
	bestScore := 0

	for _, symbol := range symbols {
		score := sp500Priority[symbol]
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

// SaveNaiveBayesModel saves model to JSON file
func SaveNaiveBayesModel(nb *NaiveBayesClassifier, path string) error {
	data := ModelData{
		PositiveWords: nb.PositiveWords,
		NeutralWords:  nb.NeutralWords,
		NegativeWords: nb.NegativeWords,
		PositiveDocs:  nb.PositiveDocs,
		NeutralDocs:   nb.NeutralDocs,
		NegativeDocs:  nb.NegativeDocs,
		VocabSize:     nb.VocabSize,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonData, 0644)
}

// LoadNaiveBayesModel loads model from JSON file
func LoadNaiveBayesModel(path string) (*NaiveBayesClassifier, error) {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data ModelData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	nb := &NaiveBayesClassifier{
		PositiveWords: data.PositiveWords,
		NeutralWords:  data.NeutralWords,
		NegativeWords: data.NegativeWords,
		PositiveDocs:  data.PositiveDocs,
		NeutralDocs:   data.NeutralDocs,
		NegativeDocs:  data.NegativeDocs,
		VocabSize:     data.VocabSize,
	}

	// Recalculate priors
	totalDocs := float64(nb.PositiveDocs + nb.NeutralDocs + nb.NegativeDocs)
	nb.PositivePrior = math.Log(float64(nb.PositiveDocs) / totalDocs)
	nb.NeutralPrior = math.Log(float64(nb.NeutralDocs) / totalDocs)
	nb.NegativePrior = math.Log(float64(nb.NegativeDocs) / totalDocs)

	return nb, nil
}
