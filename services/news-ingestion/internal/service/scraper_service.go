package service

/*
import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
)

type ScraperService interface {
	ScrapeArticleContent(ctx context.Context, articleURL string) (*ScrapedContent, error)
	ScrapeBloombergContent(ctx context.Context, articleURL string) (*ScrapedContent, error)
	ScrapeReutersContent(ctx context.Context, articleURL string) (*ScrapedContent, error)
	ScrapeGenericContent(ctx context.Context, articleURL string) (*ScrapedContent, error)
	EnrichArticleWithScrapedContent(ctx context.Context, article *model.Article) error
}

type ScrapedContent struct {
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	PublishedAt time.Time `json:"published_at"`
	ImageURL    string    `json:"image_url"`
	Summary     string    `json:"summary"`
	Keywords    []string  `json:"keywords"`
}

type scraperService struct {
	httpClient *http.Client
	userAgent  string
}

func NewScraperService() ScraperService {
	return &scraperService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		userAgent: "Mozilla/5.0 (compatible; NewsBot/1.0; +http://example.com/bot)",
	}
}

func (s *scraperService) ScrapeArticleContent(ctx context.Context, articleURL string) (*ScrapedContent, error) {
	parsedURL, err := url.Parse(articleURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Route to specialized scrapers based on domain
	switch {
	case strings.Contains(parsedURL.Host, "bloomberg.com"):
		return s.ScrapeBloombergContent(ctx, articleURL)
	case strings.Contains(parsedURL.Host, "reuters.com"):
		return s.ScrapeReutersContent(ctx, articleURL)
	default:
		return s.ScrapeGenericContent(ctx, articleURL)
	}
}

func (s *scraperService) ScrapeBloombergContent(ctx context.Context, articleURL string) (*ScrapedContent, error) {
	doc, err := s.fetchDocument(ctx, articleURL)
	if err != nil {
		return nil, err
	}

	content := &ScrapedContent{}

	// Bloomberg-specific selectors
	content.Title = s.extractText(doc, "h1[data-module='ArticleHeader']")
	if content.Title == "" {
		content.Title = s.extractText(doc, "h1")
	}

	// Extract article body
	var bodyParts []string
	doc.Find(".body-copy p").Each(func(i int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())
		if text != "" {
			bodyParts = append(bodyParts, text)
		}
	})
	content.Content = strings.Join(bodyParts, "\n\n")

	// Extract author
	content.Author = s.extractText(doc, "[data-module='ArticleHeader'] .author")

	// Extract published date
	if dateStr := s.extractText(doc, "time[data-module='ArticleHeader']"); dateStr != "" {
		if parsed, err := time.Parse("2006-01-02T15:04:05Z", dateStr); err == nil {
			content.PublishedAt = parsed
		}
	}

	// Extract image
	if img, exists := doc.Find("figure img").First().Attr("src"); exists {
		content.ImageURL = img
	}

	// Extract summary/description
	content.Summary = s.extractText(doc, "[data-module='ArticleHeader'] .summary")

	return content, nil
}

func (s *scraperService) ScrapeReutersContent(ctx context.Context, articleURL string) (*ScrapedContent, error) {
	doc, err := s.fetchDocument(ctx, articleURL)
	if err != nil {
		return nil, err
	}

	content := &ScrapedContent{}

	// Reuters-specific selectors
	content.Title = s.extractText(doc, "[data-testid='Headline']")
	if content.Title == "" {
		content.Title = s.extractText(doc, "h1")
	}

	// Extract article body
	var bodyParts []string
	doc.Find("[data-testid='paragraph'] p").Each(func(i int, sel *goquery.Selection)

*/
