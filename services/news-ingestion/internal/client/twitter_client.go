package client

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type twitterClient struct {
	bearerToken string
	baseURL     string
}

func NewTwitterClient(bearerToken, baseURL string) TwitterClient {
	if baseURL == "" {
		baseURL = "https://api.twitter.com/2"
	}

	return &twitterClient{
		bearerToken: bearerToken,
		baseURL:     baseURL,
	}
}

func (c *twitterClient) GetFinancialTweets(ctx context.Context, symbols []string) ([]*Tweet, error) {
	if c.bearerToken == "" {
		logrus.Warn("Twitter Bearer Token not configured, skipping Twitter ingestion")
		return []*Tweet{}, nil
	}

	// For now, this is a stub implementation
	// In a real implementation, you would:
	// 1. Use the Twitter API v2 to search for tweets containing financial symbols
	// 2. Handle authentication with Bearer Token
	// 3. Parse the response and convert to Tweet objects
	// 4. Handle rate limiting (Twitter has strict rate limits)

	logrus.Info("Twitter ingestion is not fully implemented yet")

	// Return empty slice for now
	return []*Tweet{}, fmt.Errorf("Twitter API integration not yet implemented")
}

// Mock implementation for testing - remove in production
func (c *twitterClient) getMockTweets(symbols []string) []*Tweet {
	tweets := []*Tweet{
		{
			ID:        "1234567890",
			Text:      "Breaking: $AAPL reports strong quarterly earnings",
			Author:    "financial_news",
			CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
			URL:       "https://twitter.com/financial_news/status/1234567890",
		},
		{
			ID:        "1234567891",
			Text:      "$TSLA stock surges on new product announcement",
			Author:    "tech_reporter",
			CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
			URL:       "https://twitter.com/tech_reporter/status/1234567891",
		},
	}

	return tweets
}
