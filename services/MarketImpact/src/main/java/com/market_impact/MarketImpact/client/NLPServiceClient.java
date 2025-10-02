// services/MarketImpact/src/main/java/com/market_impact/MarketImpact/client/NLPServiceClient.java
package com.market_impact.MarketImpact.client;

import com.market_impact.grpc.nlp.*;
import com.google.protobuf.Timestamp;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

@Component
@Slf4j
public class NLPServiceClient {

    @Value("${external.services.nlp-processing.host:localhost}")
    private String nlpServiceHost;

    @Value("${external.services.nlp-processing.port:50052}")
    private int nlpServicePort;

    private ManagedChannel channel;
    private NLPProcessingServiceGrpc.NLPProcessingServiceBlockingStub blockingStub;

    @PostConstruct
    public void init() {
        try {
            this.channel = ManagedChannelBuilder
                    .forAddress(nlpServiceHost, nlpServicePort)
                    .usePlaintext()
                    .build();

            this.blockingStub = NLPProcessingServiceGrpc.newBlockingStub(channel);
            log.info("NLP Service client initialized: {}:{}", nlpServiceHost, nlpServicePort);

            // Test connection
            testConnection();
        } catch (Exception e) {
            log.error("Failed to initialize NLP Service client: {}", e.getMessage(), e);
        }
    }

    private void testConnection() {
        try {
            log.info("Testing NLP Service connection...");
            var healthStub = io.grpc.health.v1.HealthGrpc.newBlockingStub(channel);
            var request = io.grpc.health.v1.HealthCheckRequest.newBuilder().build();
            var response = healthStub.check(request);
            log.info("NLP Service health check: status={}", response.getStatus());
        } catch (StatusRuntimeException e) {
            log.error("NLP Service connection test failed: {} - {}",
                    e.getStatus().getCode(), e.getMessage());
        }
    }

    @PreDestroy
    public void shutdown() {
        if (channel != null && !channel.isShutdown()) {
            channel.shutdown();
            log.info("NLP Service client shutdown");
        }
    }

    public Optional<SentimentData> getSentimentForArticle(String articleId) {
        try {
            log.debug("Fetching sentiment for article: {}", articleId);

            GetAnalysisRequest request = GetAnalysisRequest.newBuilder()
                    .setArticleId(articleId)
                    .build();

            GetAnalysisResponse response = blockingStub.getAnalysisResult(request);

            if (response.getFound() && response.hasResult() && response.getResult().hasSentiment()) {
                SentimentAnalysis sentiment = response.getResult().getSentiment();
                log.debug("Found sentiment for article {}: score={}, symbol={}",
                        articleId, sentiment.getCompoundScore(), sentiment.getPrimarySymbol());

                return Optional.of(SentimentData.builder()
                        .articleId(sentiment.getArticleId())
                        .compoundScore(sentiment.getCompoundScore())
                        .confidence(sentiment.getConfidence())
                        .primarySymbol(sentiment.getPrimarySymbol())
                        .analysisTimestamp(toLocalDateTime(sentiment.getAnalysisTimestamp()))
                        .build());
            }

            log.debug("No sentiment found for article: {}", articleId);
            return Optional.empty();

        } catch (StatusRuntimeException e) {
            log.error("gRPC error getting sentiment for article {}: {} - {}",
                    articleId, e.getStatus().getCode(), e.getMessage());
            return Optional.empty();
        } catch (Exception e) {
            log.error("Failed to get sentiment for article {}: {}", articleId, e.getMessage(), e);
            return Optional.empty();
        }
    }

    public List<SentimentTrendData> getSentimentTrends(String symbol, LocalDateTime startTime,
                                                       LocalDateTime endTime, int limit) {
        try {
            log.info("Fetching sentiment trends for symbol: {} from {} to {}",
                    symbol, startTime, endTime);

            SentimentTrendsRequest request = SentimentTrendsRequest.newBuilder()
                    .setSymbol(symbol)
                    .setStartTime(toTimestamp(startTime))
                    .setEndTime(toTimestamp(endTime))
                    .setInterval("hour")
                    .setLimit(limit)
                    .build();

            SentimentTrendsResponse response = blockingStub.getSentimentTrends(request);

            List<SentimentTrendData> trends = response.getTrendsList().stream()
                    .map(trend -> SentimentTrendData.builder()
                            .timestamp(toLocalDateTime(trend.getTimestamp()))
                            .symbol(trend.getSymbol())
                            .avgSentiment(trend.getAvgSentiment())
                            .articleCount(trend.getArticleCount())
                            .volatility(trend.getVolatility())
                            .build())
                    .collect(Collectors.toList());

            log.info("Retrieved {} sentiment trends for symbol: {}", trends.size(), symbol);

            if (trends.isEmpty()) {
                log.warn("No sentiment trends found for symbol: {} between {} and {}",
                        symbol, startTime, endTime);
            } else {
                log.debug("First trend for {}: avgSentiment={}, articleCount={}",
                        symbol, trends.get(0).getAvgSentiment(), trends.get(0).getArticleCount());
            }

            return trends;

        } catch (StatusRuntimeException e) {
            log.error("gRPC error getting sentiment trends for {}: {} - {}",
                    symbol, e.getStatus().getCode(), e.getMessage());
            return new ArrayList<>();
        } catch (Exception e) {
            log.error("Failed to get sentiment trends for {}: {}", symbol, e.getMessage(), e);
            return new ArrayList<>();
        }
    }

    private LocalDateTime toLocalDateTime(Timestamp timestamp) {
        Instant instant = Instant.ofEpochSecond(timestamp.getSeconds(), timestamp.getNanos());
        return LocalDateTime.ofInstant(instant, ZoneOffset.UTC);
    }

    private Timestamp toTimestamp(LocalDateTime dateTime) {
        Instant instant = dateTime.toInstant(ZoneOffset.UTC);
        return Timestamp.newBuilder()
                .setSeconds(instant.getEpochSecond())
                .setNanos(instant.getNano())
                .build();
    }

  public boolean isConnected() {
      try {
          var healthStub = io.grpc.health.v1.HealthGrpc.newBlockingStub(channel);
          var request = io.grpc.health.v1.HealthCheckRequest.newBuilder().build();
          healthStub.check(request);
          return true;
      } catch (Exception e) {
          return false;
      }
  }
  }