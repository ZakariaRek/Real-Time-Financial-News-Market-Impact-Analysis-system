package client

import (
	"context"
	"time"
)

// NewsAPIClient handles NewsAPI integration
type NewsAPIClient interface {
	GetFinancialNews(ctx context.Context, query string) ([]*NewsAPIArticle, error)
	GetTopHeadlines(ctx context.Context, category string) ([]*NewsAPIArticle, error)
}

// RSSClient handles RSS feed parsing
type RSSClient interface {
	FetchFeed(ctx context.Context, url string) ([]*RSSItem, error)
}

// TwitterClient handles Twitter API integration
type TwitterClient interface {
	GetFinancialTweets(ctx context.Context, symbols []string) ([]*Tweet, error)
}

// NewsAPIArticle represents an article from NewsAPI
type NewsAPIArticle struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Source      struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"source"`
	Author string `json:"author"`
}

// RSSItem represents an RSS feed item
type RSSItem struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	PubDate     time.Time `json:"pubDate"`
	GUID        string    `json:"guid"`
}

// Tweet represents a Twitter post
type Tweet struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	URL       string    `json:"url"`
}
