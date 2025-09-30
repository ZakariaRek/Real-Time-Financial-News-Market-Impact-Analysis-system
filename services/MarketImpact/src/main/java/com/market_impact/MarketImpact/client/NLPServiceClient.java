package com.market_impact.MarketImpact.client;

import com.market_impact.grpc.nlp.*;
import com.google.protobuf.Timestamp;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;


import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
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
        this.channel = ManagedChannelBuilder
                .forAddress(nlpServiceHost, nlpServicePort)
                .usePlaintext()
                .build();

        this.blockingStub = NLPProcessingServiceGrpc.newBlockingStub(channel);
        log.info("NLP Service client initialized: {}:{}", nlpServiceHost, nlpServicePort);
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
            GetAnalysisRequest request = GetAnalysisRequest.newBuilder()
                    .setArticleId(articleId)
                    .build();

            GetAnalysisResponse response = blockingStub.getAnalysisResult(request);

            if (response.getFound() && response.hasResult() && response.getResult().hasSentiment()) {
                SentimentAnalysis sentiment = response.getResult().getSentiment();
                return Optional.of(SentimentData.builder()
                        .articleId(sentiment.getArticleId())
                        .compoundScore(sentiment.getCompoundScore())
                        .confidence(sentiment.getConfidence())
                        .primarySymbol(sentiment.getPrimarySymbol())
                        .analysisTimestamp(toLocalDateTime(sentiment.getAnalysisTimestamp()))
                        .build());
            }

            return Optional.empty();
        } catch (Exception e) {
            log.error("Failed to get sentiment for article {}: {}", articleId, e.getMessage());
            return Optional.empty();
        }
    }

    public List<SentimentTrendData> getSentimentTrends(String symbol, LocalDateTime startTime, LocalDateTime endTime, int limit) {
        try {
            SentimentTrendsRequest request = SentimentTrendsRequest.newBuilder()
                    .setSymbol(symbol)
                    .setStartTime(toTimestamp(startTime))
                    .setEndTime(toTimestamp(endTime))
                    .setInterval("hour")
                    .setLimit(limit)
                    .build();

            SentimentTrendsResponse response = blockingStub.getSentimentTrends(request);

            return response.getTrendsList().stream()
                    .map(trend -> SentimentTrendData.builder()
                            .timestamp(toLocalDateTime(trend.getTimestamp()))
                            .symbol(trend.getSymbol())
                            .avgSentiment(trend.getAvgSentiment())
                            .articleCount(trend.getArticleCount())
                            .volatility(trend.getVolatility())
                            .build())
                    .collect(Collectors.toList());
        } catch (Exception e) {
            log.error("Failed to get sentiment trends for {}: {}", symbol, e.getMessage());
            return List.of();
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
}