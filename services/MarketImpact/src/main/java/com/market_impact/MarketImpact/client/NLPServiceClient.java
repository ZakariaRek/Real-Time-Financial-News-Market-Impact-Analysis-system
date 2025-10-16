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
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.TimeUnit;
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
    private volatile boolean connected = false;
    private volatile int consecutiveFailures = 0;
    private static final int MAX_CONSECUTIVE_FAILURES = 3;

    @PostConstruct
    public void init() {
        log.info("🔌 Initializing NLP Service client");
        log.info("📍 Target: {}:{}", nlpServiceHost, nlpServicePort);

        initializeChannel();
        log.info("✅ NLP Service client initialized (connection will be verified asynchronously)");
    }

    private void initializeChannel() {
        try {
            if (channel != null && !channel.isShutdown()) {
                channel.shutdown();
                try {
                    channel.awaitTermination(5, TimeUnit.SECONDS);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                }
            }

            // ✅ FIX: Add message size configuration
            this.channel = ManagedChannelBuilder
                    .forAddress(nlpServiceHost, nlpServicePort)
                    .usePlaintext()
                    .maxInboundMessageSize(4 * 1024 * 1024)      // ✅ 4MB - matches NLP service
                    .maxInboundMetadataSize(8 * 1024)            // ✅ 8KB
                    .keepAliveTime(30, TimeUnit.SECONDS)
                    .keepAliveTimeout(10, TimeUnit.SECONDS)
                    .keepAliveWithoutCalls(true)
                    .idleTimeout(60, TimeUnit.SECONDS)
                    .build();

            this.blockingStub = NLPProcessingServiceGrpc.newBlockingStub(channel)
                    .withDeadlineAfter(30, TimeUnit.SECONDS);

            log.info("✅ NLP Service channel created: {}:{}", nlpServiceHost, nlpServicePort);
        } catch (Exception e) {
            log.error("❌ Failed to create NLP Service channel: {}", e.getMessage());
            connected = false;
        }
    }
    /**
     * Periodically test connection and reconnect if needed
     */
    @Scheduled(fixedDelay = 30000, initialDelay = 5000) // Every 30 seconds, start after 5 seconds
    public void checkConnection() {
        try {
            if (channel == null || channel.isShutdown() || channel.isTerminated()) {
                log.warn("⚠️ Channel is shutdown/terminated, reinitializing...");
                initializeChannel();
            }

            // Test connection
            var healthStub = io.grpc.health.v1.HealthGrpc.newBlockingStub(channel)
                    .withDeadlineAfter(5, TimeUnit.SECONDS);
            var request = io.grpc.health.v1.HealthCheckRequest.newBuilder().build();
            var response = healthStub.check(request);

            if (!connected) {
                log.info("✅ NLP Service connection established: status={}", response.getStatus());
            }
            connected = true;
            consecutiveFailures = 0;

        } catch (StatusRuntimeException e) {
            consecutiveFailures++;

            if (connected || consecutiveFailures % 10 == 1) { // Log every 10th failure to avoid spam
                log.warn("⚠️ NLP Service health check failed (attempt {}/{}): {} - {}",
                        consecutiveFailures, MAX_CONSECUTIVE_FAILURES,
                        e.getStatus().getCode(), e.getMessage());
            }

            if (consecutiveFailures >= MAX_CONSECUTIVE_FAILURES) {
                connected = false;
                if (consecutiveFailures == MAX_CONSECUTIVE_FAILURES) {
                    log.error("❌ NLP Service marked as disconnected after {} consecutive failures",
                            MAX_CONSECUTIVE_FAILURES);
                }
            }

            // Try to reinitialize channel after multiple failures
            if (consecutiveFailures % 5 == 0) {
                log.info("🔄 Attempting to reinitialize channel after {} failures", consecutiveFailures);
                initializeChannel();
            }
        } catch (Exception e) {
            consecutiveFailures++;
            if (connected || consecutiveFailures % 10 == 1) {
                log.error("⚠️ Unexpected error during health check: {}", e.getMessage());
            }
            connected = false;
        }
    }

    @PreDestroy
    public void shutdown() {
        if (channel != null && !channel.isShutdown()) {
            try {
                channel.shutdown();
                if (!channel.awaitTermination(5, TimeUnit.SECONDS)) {
                    channel.shutdownNow();
                }
                log.info("🔌 NLP Service client shutdown");
            } catch (InterruptedException e) {
                log.warn("Interrupted during shutdown");
                channel.shutdownNow();
                Thread.currentThread().interrupt();
            }
        }
    }

    public Optional<SentimentData> getSentimentForArticle(String articleId) {
        if (!isConnected()) {
            log.warn("⚠️ NLP Service not connected, skipping sentiment request for article {}", articleId);
            return Optional.empty();
        }

        try {
            log.debug("📤 Fetching sentiment for article: {}", articleId);

            GetAnalysisRequest request = GetAnalysisRequest.newBuilder()
                    .setArticleId(articleId)
                    .build();

            GetAnalysisResponse response = blockingStub.getAnalysisResult(request);

            if (response.getFound() && response.hasResult() && response.getResult().hasSentiment()) {
                SentimentAnalysis sentiment = response.getResult().getSentiment();
                log.debug("✅ Found sentiment for article {}: score={}, symbol={}",
                        articleId, sentiment.getCompoundScore(), sentiment.getPrimarySymbol());

                return Optional.of(SentimentData.builder()
                        .articleId(sentiment.getArticleId())
                        .compoundScore(sentiment.getCompoundScore())
                        .confidence(sentiment.getConfidence())
                        .primarySymbol(sentiment.getPrimarySymbol())
                        .analysisTimestamp(toLocalDateTime(sentiment.getAnalysisTimestamp()))
                        .build());
            }

            log.debug("ℹ️ No sentiment found for article: {}", articleId);
            return Optional.empty();

        } catch (StatusRuntimeException e) {
            log.error("❌ gRPC error getting sentiment for article {}: {} - {}",
                    articleId, e.getStatus().getCode(), e.getMessage());

            // Mark as disconnected on certain errors
            if (e.getStatus().getCode() == io.grpc.Status.Code.UNAVAILABLE ||
                    e.getStatus().getCode() == io.grpc.Status.Code.DEADLINE_EXCEEDED) {
                consecutiveFailures++;
                if (consecutiveFailures >= MAX_CONSECUTIVE_FAILURES) {
                    connected = false;
                }
            }
            return Optional.empty();
        } catch (Exception e) {
            log.error("❌ Failed to get sentiment for article {}: {}", articleId, e.getMessage(), e);
            return Optional.empty();
        }
    }

    public List<SentimentTrendData> getSentimentTrends(String symbol, LocalDateTime startTime,
                                                       LocalDateTime endTime, int limit) {
        if (!isConnected()) {
            log.warn("⚠️ NLP Service not connected, skipping sentiment trends request for symbol {}", symbol);
            return new ArrayList<>();
        }

        try {
            log.info("📤 Fetching sentiment trends for symbol: {} from {} to {}",
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

            log.info("✅ Retrieved {} sentiment trends for symbol: {}", trends.size(), symbol);

            if (trends.isEmpty()) {
                log.warn("⚠️ No sentiment trends found for symbol: {} between {} and {}",
                        symbol, startTime, endTime);
            } else {
                log.debug("📊 First trend for {}: avgSentiment={}, articleCount={}",
                        symbol, trends.get(0).getAvgSentiment(), trends.get(0).getArticleCount());
            }

            return trends;

        } catch (StatusRuntimeException e) {
            log.error("❌ gRPC error getting sentiment trends for {}: {} - {}",
                    symbol, e.getStatus().getCode(), e.getMessage());

            if (e.getStatus().getCode() == io.grpc.Status.Code.UNAVAILABLE ||
                    e.getStatus().getCode() == io.grpc.Status.Code.DEADLINE_EXCEEDED) {
                consecutiveFailures++;
                if (consecutiveFailures >= MAX_CONSECUTIVE_FAILURES) {
                    connected = false;
                }
            }
            return new ArrayList<>();
        } catch (Exception e) {
            log.error("❌ Failed to get sentiment trends for {}: {}", symbol, e.getMessage(), e);
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
        return connected && channel != null && !channel.isShutdown() && !channel.isTerminated();
    }
}