// services/news-ingestion/internal/client/nlp_client.go
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	
	"google.golang.org/protobuf/types/known/timestamppb" // ✅ Add this

	"github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/news-ingestion/internal/model"
	nlpv1 "github.com/ZakariaRek/Real-Time-Financial-News-Market-Impact-Analysis-system/services/nlp-processing/proto/gen/nlp/v1"
)

// NLPProcessingClient interface for NLP service communication
type NLPProcessingClient interface {
	ProcessBatch(ctx context.Context, articles []*model.Article) error
	ProcessSingle(ctx context.Context, article *model.Article) error
	GetProcessingStatus(ctx context.Context) (*ProcessingStatus, error)
	Close() error
}

// ProcessingStatus represents the status of NLP processing
type ProcessingStatus struct {
	PendingArticles    int32
	ProcessingArticles int32
	CompletedToday     int32
	FailedToday        int32
}

type nlpClient struct {
	conn   *grpc.ClientConn
	client nlpv1.NLPProcessingServiceClient
}

// NewNLPProcessingClient creates a new NLP processing client
func NewNLPProcessingClient(endpoint string) (NLPProcessingClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logrus.Infof("Connecting to NLP service at %s", endpoint)

	conn, err := grpc.DialContext(ctx, endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NLP service: %w", err)
	}

	client := nlpv1.NewNLPProcessingServiceClient(conn)

	logrus.Info("Successfully connected to NLP service")

	return &nlpClient{
		conn:   conn,
		client: client,
	}, nil
}

// ProcessBatch sends a batch of articles to NLP service for processing
func (c *nlpClient) ProcessBatch(ctx context.Context, articles []*model.Article) error {
	if len(articles) == 0 {
		return nil
	}

	// Convert model.Article to proto.Article
	protoArticles := make([]*nlpv1.Article, len(articles))
	for i, article := range articles {
		protoArticles[i] = c.modelToProto(article)
	}

	// Create batch request
	req := &nlpv1.BatchProcessRequest{
		Articles: protoArticles,
		Options: &nlpv1.ProcessingOptions{
			EnableSentiment:     true,
			EnableNer:           false, // Disabled in simplified version
			EnableTopic:         false, // Disabled in simplified version
			ConfidenceThreshold: 0.3,
			ModelVersion:        "1.0",
		},
	}

	// Send batch to NLP service
	resp, err := c.client.ProcessBatch(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to process batch: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"successful": resp.SuccessfulCount,
		"failed":     resp.FailedCount,
		"total":      len(articles),
	}).Info("Batch processing completed")

	if resp.FailedCount > 0 {
		logrus.WithField("errors", resp.Errors).Warn("Some articles failed processing")
	}

	return nil
}

// ProcessSingle sends a single article to NLP service for processing
func (c *nlpClient) ProcessSingle(ctx context.Context, article *model.Article) error {
	protoArticle := c.modelToProto(article)

	req := &nlpv1.ProcessArticleRequest{
		Article: protoArticle,
		Options: &nlpv1.ProcessingOptions{
			EnableSentiment:     true,
			ConfidenceThreshold: 0.3,
			ModelVersion:        "1.0",
		},
	}

	resp, err := c.client.ProcessArticle(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to process article: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("processing failed: %s", resp.Message)
	}

	logrus.WithField("article_id", article.ID).Debug("Article processed successfully")
	return nil
}

// GetProcessingStatus gets the current processing status from NLP service
func (c *nlpClient) GetProcessingStatus(ctx context.Context) (*ProcessingStatus, error) {
	resp, err := c.client.GetProcessingStatus(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get processing status: %w", err)
	}

	return &ProcessingStatus{
		PendingArticles:    resp.PendingArticles,
		ProcessingArticles: resp.ProcessingArticles,
		CompletedToday:     resp.CompletedToday,
		FailedToday:        resp.FailedToday,
	}, nil
}

// Close closes the gRPC connection
func (c *nlpClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// modelToProto converts model.Article to proto.Article
func (c *nlpClient) modelToProto(article *model.Article) *nlpv1.Article {
	return &nlpv1.Article{
		Id:          article.ID.String(),
		Title:       article.Title,
		Content:     article.Content,
		Url:         article.URL,
		Symbols:     article.Symbols,
		SourceId:    uint32(article.SourceID),
		PublishedAt: timestamppb.New(article.PublishedAt),
	}
}
