// services/nlp-processing/internal/service/ner_service.go
package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/internal/model"
)

type NERService interface {
	ExtractEntities(ctx context.Context, articleID, text string) ([]model.EntityRecognition, error)
	BatchExtractEntities(ctx context.Context, articles []model.AnalysisRequest) (map[string][]model.EntityRecognition, error)
	LoadModel(ctx context.Context) error
	IsModelLoaded() bool
}

type nerService struct {
	modelLoaded         bool
	modelPath           string
	spacyModel          string
	confidenceThreshold float32

	// Placeholder for actual spaCy/NER model
	// In production, this would be replaced with actual model interface
	// spacyClient *spacy.Client
}

type NERConfig struct {
	ModelPath           string  `mapstructure:"model_path"`
	SpacyModel          string  `mapstructure:"spacy_model"`
	ConfidenceThreshold float32 `mapstructure:"confidence_threshold"`
}

func NewNERService(config NERConfig) NERService {
	return &nerService{
		modelLoaded:         false,
		modelPath:           config.ModelPath,
		spacyModel:          config.SpacyModel,
		confidenceThreshold: config.ConfidenceThreshold,
	}
}

func (n *nerService) LoadModel(ctx context.Context) error {
	logrus.Info("Loading SpaCy NER model...")

	// TODO: Replace with actual model loading when models are provided
	// Example implementation structure:
	/*
		client, err := spacy.NewClient(n.spacyModel, n.modelPath)
		if err != nil {
			return fmt.Errorf("failed to load SpaCy NER model: %w", err)
		}
		n.spacyClient = client
	*/

	// Simulate model loading time
	time.Sleep(1 * time.Second)

	n.modelLoaded = true
	logrus.Info("SpaCy NER model loaded successfully")

	return nil
}

func (n *nerService) IsModelLoaded() bool {
	return n.modelLoaded
}

func (n *nerService) ExtractEntities(ctx context.Context, articleID, text string) ([]model.EntityRecognition, error) {
	if !n.modelLoaded {
		return nil, fmt.Errorf("NER model not loaded")
	}

	// TODO: Replace with actual model inference when models are provided
	// Example implementation structure:
	/*
		entities, err := n.spacyClient.ExtractEntities(text)
		if err != nil {
			return nil, fmt.Errorf("NER extraction failed: %w", err)
		}
	*/

	// Stub implementation - extract entities using pattern matching
	entities := n.stubEntityExtraction(articleID, text)

	// Filter entities by confidence threshold
	filteredEntities := make([]model.EntityRecognition, 0)
	for _, entity := range entities {
		if entity.Confidence >= n.confidenceThreshold {
			filteredEntities = append(filteredEntities, entity)
		}
	}

	logrus.WithFields(logrus.Fields{
		"article_id":        articleID,
		"total_entities":    len(entities),
		"filtered_entities": len(filteredEntities),
	}).Debug("Named entity recognition completed")

	return filteredEntities, nil
}

func (n *nerService) BatchExtractEntities(ctx context.Context, articles []model.AnalysisRequest) (map[string][]model.EntityRecognition, error) {
	if !n.modelLoaded {
		return nil, fmt.Errorf("NER model not loaded")
	}

	results := make(map[string][]model.EntityRecognition)

	for _, article := range articles {
		entities, err := n.ExtractEntities(ctx, article.ArticleID, article.Title+" "+article.Content)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to extract entities for article %s", article.ArticleID)
			continue
		}
		results[article.ArticleID] = entities
	}

	return results, nil
}

// Stub implementation for entity extraction until actual models are provided
func (n *nerService) stubEntityExtraction(articleID, text string) []model.EntityRecognition {
	entities := make([]model.EntityRecognition, 0)

	// Extract stock symbols (pattern: $SYMBOL or uppercase 2-5 letters)
	stockSymbols := n.extractStockSymbols(text)
	for _, symbol := range stockSymbols {
		entity := model.EntityRecognition{
			ArticleID:      articleID,
			EntityText:     symbol.Text,
			EntityType:     "STOCK_SYMBOL",
			StockSymbol:    symbol.Symbol,
			Confidence:     symbol.Confidence,
			EntityCategory: "Financial_Instrument",
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
		entities = append(entities, entity)
	}

	// Extract company names (simplified pattern matching)
	companies := n.extractCompanies(text)
	for _, company := range companies {
		entity := model.EntityRecognition{
			ArticleID:      articleID,
			EntityText:     company.Text,
			EntityType:     "ORG",
			StockSymbol:    company.Symbol,
			Confidence:     company.Confidence,
			EntityCategory: "Company",
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
		entities = append(entities, entity)
	}

	// Extract person names (simplified pattern matching)
	persons := n.extractPersons(text)
	for _, person := range persons {
		entity := model.EntityRecognition{
			ArticleID:      articleID,
			EntityText:     person.Text,
			EntityType:     "PERSON",
			Confidence:     person.Confidence,
			EntityCategory: "Executive",
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
		entities = append(entities, entity)
	}

	// Extract monetary amounts
	moneyAmounts := n.extractMoneyAmounts(text)
	for _, money := range moneyAmounts {
		entity := model.EntityRecognition{
			ArticleID:      articleID,
			EntityText:     money.Text,
			EntityType:     "MONEY",
			Confidence:     money.Confidence,
			EntityCategory: "Financial_Amount",
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
		entities = append(entities, entity)
	}

	return entities
}

type ExtractedEntity struct {
	Text       string
	Symbol     string
	Confidence float32
}

func (n *nerService) extractStockSymbols(text string) []ExtractedEntity {
	entities := make([]ExtractedEntity, 0)

	// Pattern for $SYMBOL format
	dollarPattern := regexp.MustCompile(`\$[A-Z]{2,5}`)
	matches := dollarPattern.FindAllString(text, -1)

	for _, match := range matches {
		symbol := strings.TrimPrefix(match, "$")
		entity := ExtractedEntity{
			Text:       match,
			Symbol:     symbol,
			Confidence: 0.9, // High confidence for explicit symbol format
		}
		entities = append(entities, entity)
	}

	// Pattern for common stock symbols without $
	knownSymbols := []string{"AAPL", "GOOGL", "MSFT", "AMZN", "TSLA", "META", "NVDA", "AMD", "NFLX", "CRM"}
	for _, symbol := range knownSymbols {
		pattern := regexp.MustCompile(`\b` + symbol + `\b`)
		if pattern.MatchString(text) {
			entity := ExtractedEntity{
				Text:       symbol,
				Symbol:     symbol,
				Confidence: 0.8,
			}
			entities = append(entities, entity)
		}
	}

	return entities
}

func (n *nerService) extractCompanies(text string) []ExtractedEntity {
	entities := make([]ExtractedEntity, 0)

	// Known company mappings
	companyMap := map[string]string{
		"Apple":      "AAPL",
		"Google":     "GOOGL",
		"Alphabet":   "GOOGL",
		"Microsoft":  "MSFT",
		"Amazon":     "AMZN",
		"Tesla":      "TSLA",
		"Meta":       "META",
		"Facebook":   "META",
		"NVIDIA":     "NVDA",
		"AMD":        "AMD",
		"Netflix":    "NFLX",
		"Salesforce": "CRM",
	}

	for company, symbol := range companyMap {
		pattern := regexp.MustCompile(`\b` + company + `\b`)
		if pattern.MatchString(text) {
			entity := ExtractedEntity{
				Text:       company,
				Symbol:     symbol,
				Confidence: 0.85,
			}
			entities = append(entities, entity)
		}
	}

	// Pattern for "Inc.", "Corp.", "Ltd.", etc.
	corpPattern := regexp.MustCompile(`[A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\s+(?:Inc\.|Corp\.|Ltd\.|LLC|Co\.)`)
	matches := corpPattern.FindAllString(text, -1)

	for _, match := range matches {
		entity := ExtractedEntity{
			Text:       match,
			Confidence: 0.7,
		}
		entities = append(entities, entity)
	}

	return entities
}

func (n *nerService) extractPersons(text string) []ExtractedEntity {
	entities := make([]ExtractedEntity, 0)

	// Known executive names
	executives := []string{"Elon Musk", "Tim Cook", "Sundar Pichai", "Satya Nadella", "Jeff Bezos", "Mark Zuckerberg"}

	for _, executive := range executives {
		if strings.Contains(text, executive) {
			entity := ExtractedEntity{
				Text:       executive,
				Confidence: 0.9,
			}
			entities = append(entities, entity)
		}
	}

	// Pattern for potential names (simplified)
	namePattern := regexp.MustCompile(`\b[A-Z][a-z]+\s+[A-Z][a-z]+\b`)
	matches := namePattern.FindAllString(text, -1)

	for _, match := range matches {
		// Skip common false positives
		if n.isLikelyName(match) {
			entity := ExtractedEntity{
				Text:       match,
				Confidence: 0.6, // Lower confidence for pattern-based detection
			}
			entities = append(entities, entity)
		}
	}

	return entities
}

func (n *nerService) extractMoneyAmounts(text string) []ExtractedEntity {
	entities := make([]ExtractedEntity, 0)

	// Pattern for monetary amounts: $X.X million/billion/trillion
	moneyPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\$\d+(?:\.\d+)?\s*(?:million|billion|trillion)`),
		regexp.MustCompile(`\$\d+(?:,\d{3})*(?:\.\d{2})?`),
		regexp.MustCompile(`\d+(?:\.\d+)?\s*(?:million|billion|trillion)\s*dollars?`),
	}

	for _, pattern := range moneyPatterns {
		matches := pattern.FindAllString(text, -1)
		for _, match := range matches {
			entity := ExtractedEntity{
				Text:       match,
				Confidence: 0.8,
			}
			entities = append(entities, entity)
		}
	}

	return entities
}

func (n *nerService) isLikelyName(candidate string) bool {
	// Filter out common false positives
	falsePositives := []string{
		"New York", "San Francisco", "Los Angeles", "United States",
		"Wall Street", "Stock Market", "Chief Executive", "Chief Financial",
	}

	for _, fp := range falsePositives {
		if strings.EqualFold(candidate, fp) {
			return false
		}
	}

	return true
}
