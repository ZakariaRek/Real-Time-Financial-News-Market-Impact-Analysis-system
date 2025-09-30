// services/news-ingestion/internal/database/seed.go
package database

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
)

// SeedData seeds initial data into the database
func (d *Database) SeedData(ctx context.Context) error {
	logrus.Info("Checking if database needs seeding...")

	// Check if sources already exist
	var count int64
	if err := d.DB.Model(&model.NewsSource{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		logrus.Infof("Database already contains %d sources, skipping seed", count)
		return nil
	}

	logrus.Info("Seeding initial news sources...")

	// Get RSS feeds from config
	rssFeedsConfig := viper.Get("news_sources.rss_feeds")

	sources := []model.NewsSource{}

	// Add RSS sources from config
	if rssFeeds, ok := rssFeedsConfig.([]interface{}); ok {
		for _, feed := range rssFeeds {
			if feedMap, ok := feed.(map[string]interface{}); ok {
				enabled := true
				if enabledVal, exists := feedMap["enabled"]; exists {
					enabled = enabledVal.(bool)
				}

				if !enabled {
					continue
				}

				name := feedMap["name"].(string)
				url := feedMap["url"].(string)

				sources = append(sources, model.NewsSource{
					Name:               name,
					SourceType:         "RSS",
					BaseURL:            url,
					RateLimitPerMinute: 10,
					Status:             "active",
					SuccessRate:        0.0,
				})
			}
		}
	}

	// Insert sources
	if len(sources) == 0 {
		logrus.Warn("No valid news sources configured. Please configure RSS feeds in config.yaml")
		return nil
	}

	for _, source := range sources {
		if err := d.DB.Create(&source).Error; err != nil {
			logrus.WithError(err).Errorf("Failed to seed source: %s", source.Name)
			continue
		}
		logrus.Infof("✓ Seeded source: %s (%s) - %s", source.Name, source.SourceType, source.BaseURL)
	}

	logrus.Infof("✓ Successfully seeded %d news sources", len(sources))
	return nil
}
