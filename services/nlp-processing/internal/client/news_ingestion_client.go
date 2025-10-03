package client

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	newsv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/proto/services/news-ingestion/proto/gen"
)

type NewsIngestionClient struct {
	conn   *grpc.ClientConn
	client newsv1.NewsServiceClient
}

func NewNewsIngestionClient(endpoint string) (*NewsIngestionClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to news-ingestion service: %w", err)
	}

	client := newsv1.NewNewsServiceClient(conn)

	return &NewsIngestionClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *NewsIngestionClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// StreamPendingArticles streams pending articles for NLP processing
func (c *NewsIngestionClient) StreamPendingArticles(ctx context.Context, batchSize int32) (<-chan *newsv1.Article, <-chan error) {
	articleChan := make(chan *newsv1.Article, 100)
	errorChan := make(chan error, 1)

	go func() {
		defer close(articleChan)
		defer close(errorChan)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := c.fetchAndStreamBatch(ctx, batchSize, articleChan); err != nil {
					if err != io.EOF {
						errorChan <- err
						logrus.WithError(err).Error("Error streaming articles")
					}
					time.Sleep(5 * time.Second) // Wait before retry
					continue
				}
			}
		}
	}()

	return articleChan, errorChan
}

func (c *NewsIngestionClient) fetchAndStreamBatch(ctx context.Context, batchSize int32, articleChan chan<- *newsv1.Article) error {
	stream, err := c.client.StreamArticles(ctx, &newsv1.StreamArticlesRequest{
		Status:     "pending",
		BatchSize:  batchSize,
		Continuous: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	for {
		article, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to receive article: %w", err)
		}

		select {
		case articleChan <- article.Article:
			logrus.WithField("article_id", article.Article.Id).Debug("Received article for processing")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// AcknowledgeProcessing acknowledges that an article has been processed
func (c *NewsIngestionClient) AcknowledgeProcessing(ctx context.Context, articleID string, success bool, errorMsg string, processingTimeMs int64) error {
	_, err := c.client.AcknowledgeArticleProcessing(ctx, &newsv1.AcknowledgeArticleProcessingRequest{
		ArticleId:        articleID,
		Success:          success,
		ErrorMessage:     errorMsg,
		ProcessingTimeMs: processingTimeMs,
	})
	return err
}

// GetPendingArticlesCount gets count of pending articles
func (c *NewsIngestionClient) GetPendingArticlesCount(ctx context.Context) (int32, error) {
	resp, err := c.client.ListArticles(ctx, &newsv1.ListArticlesRequest{
		Status: "pending",
		Limit:  1000,
	})
	if err != nil {
		return 0, err
	}
	return resp.TotalCount, nil
}
