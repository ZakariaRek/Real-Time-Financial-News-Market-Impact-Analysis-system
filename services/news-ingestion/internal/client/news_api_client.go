// services/news-ingestion/internal/client/news_api_client.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

type newsAPIClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

type NewsAPIResponse struct {
	Status       string            `json:"status"`
	TotalResults int               `json:"totalResults"`
	Articles     []*NewsAPIArticle `json:"articles"`
	Code         string            `json:"code,omitempty"`
	Message      string            `json:"message,omitempty"`
}

func NewNewsAPIClient(apiKey, baseURL string) NewsAPIClient {
	if baseURL == "" {
		baseURL = "https://newsapi.org/v2"
	}

	return &newsAPIClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

func (c *newsAPIClient) GetFinancialNews(ctx context.Context, query string) ([]*NewsAPIArticle, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("NewsAPI key is required")
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("language", "en")
	params.Set("sortBy", "publishedAt")
	params.Set("pageSize", "100")
	params.Set("apiKey", c.apiKey)

	url := fmt.Sprintf("%s/everything?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "NewsBot/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from NewsAPI: %w", err)
	}
	defer resp.Body.Close()

	var apiResp NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode NewsAPI response: %w", err)
	}

	if apiResp.Status != "ok" {
		return nil, fmt.Errorf("NewsAPI error: %s - %s", apiResp.Code, apiResp.Message)
	}

	// Filter and clean articles
	var articles []*NewsAPIArticle
	for _, article := range apiResp.Articles {
		// Skip articles without title or URL
		if article.Title == "" || article.URL == "" {
			continue
		}

		// Skip articles marked as removed
		if article.Title == "[Removed]" {
			continue
		}

		articles = append(articles, article)
	}

	logrus.Infof("Successfully fetched %d articles from NewsAPI for query: %s", len(articles), query)
	return articles, nil
}

func (c *newsAPIClient) GetTopHeadlines(ctx context.Context, category string) ([]*NewsAPIArticle, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("NewsAPI key is required")
	}

	params := url.Values{}
	params.Set("category", category)
	params.Set("language", "en")
	params.Set("pageSize", "100")
	params.Set("apiKey", c.apiKey)

	url := fmt.Sprintf("%s/top-headlines?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "NewsBot/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from NewsAPI: %w", err)
	}
	defer resp.Body.Close()

	var apiResp NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode NewsAPI response: %w", err)
	}

	if apiResp.Status != "ok" {
		return nil, fmt.Errorf("NewsAPI error: %s - %s", apiResp.Code, apiResp.Message)
	}

	// Filter and clean articles
	var articles []*NewsAPIArticle
	for _, article := range apiResp.Articles {
		// Skip articles without title or URL
		if article.Title == "" || article.URL == "" {
			continue
		}

		// Skip articles marked as removed
		if article.Title == "[Removed]" {
			continue
		}

		articles = append(articles, article)
	}

	logrus.Infof("Successfully fetched %d top headlines from NewsAPI for category: %s", len(articles), category)
	return articles, nil
}
