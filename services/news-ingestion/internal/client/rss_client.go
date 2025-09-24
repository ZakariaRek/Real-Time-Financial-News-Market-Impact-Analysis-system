package client

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type rssClient struct {
	httpClient *http.Client
	userAgent  string
}

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

func NewRSSClient() RSSClient {
	return &rssClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "Mozilla/5.0 (compatible; NewsBot/1.0; +http://example.com/bot)",
	}
}

func (c *rssClient) FetchFeed(ctx context.Context, url string) ([]*RSSItem, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS feed returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	var items []*RSSItem
	for _, item := range rss.Channel.Items {
		pubDate, err := c.parseRSSDate(item.PubDate)
		if err != nil {
			logrus.WithError(err).Warnf("Failed to parse date for RSS item: %s", item.Title)
			pubDate = time.Now().UTC()
		}

		rssItem := &RSSItem{
			Title:       strings.TrimSpace(item.Title),
			Description: strings.TrimSpace(item.Description),
			Link:        strings.TrimSpace(item.Link),
			PubDate:     pubDate,
			GUID:        strings.TrimSpace(item.GUID),
		}

		// Skip items with empty title or link
		if rssItem.Title == "" || rssItem.Link == "" {
			continue
		}

		items = append(items, rssItem)
	}

	logrus.Infof("Successfully fetched %d items from RSS feed: %s", len(items), url)
	return items, nil
}

func (c *rssClient) parseRSSDate(dateStr string) (time.Time, error) {
	// Common RSS date formats
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"Mon, 02 Jan 2006 15:04:05 MST",
	}

	dateStr = strings.TrimSpace(dateStr)

	for _, format := range formats {
		if parsed, err := time.Parse(format, dateStr); err == nil {
			return parsed.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
