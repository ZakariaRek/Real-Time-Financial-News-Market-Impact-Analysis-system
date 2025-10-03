// services/MarketImpact/src/main/java/com/market_impact/MarketImpact/scheduler/SentimentProcessingScheduler.java
package com.market_impact.MarketImpact.scheduler;

import com.market_impact.MarketImpact.Services.MarketImpactPredictionService;
import com.market_impact.MarketImpact.Services.SP500MarketImpactService;
import com.market_impact.MarketImpact.client.NLPServiceClient;
import com.market_impact.MarketImpact.client.SentimentTrendData;
import com.market_impact.MarketImpact.Config.SP500Config;
import com.market_impact.MarketImpact.entity.MarketPrediction;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;

@Component
@RequiredArgsConstructor
@Slf4j
public class SentimentProcessingScheduler {

    private final NLPServiceClient nlpServiceClient;
    private final MarketImpactPredictionService predictionService;
    private final SP500MarketImpactService sp500Service;
    private final SP500Config sp500Config;

    private final ExecutorService executorService = Executors.newFixedThreadPool(5);

    private final AtomicInteger successCount = new AtomicInteger(0);
    private final AtomicInteger failureCount = new AtomicInteger(0);
    private final AtomicInteger noDataCount = new AtomicInteger(0);

    @EventListener(ApplicationReadyEvent.class)
    public void onApplicationReady() {
        log.info("Application started - checking NLP service connection");
        if (nlpServiceClient.isConnected()) {
            log.info("NLP Service is connected and ready");
            // Run initial processing after 30 seconds
            CompletableFuture.delayedExecutor(30, java.util.concurrent.TimeUnit.SECONDS)
                    .execute(this::processSP500Sentiment);
        } else {
            log.error("NLP Service is NOT connected!");
        }
    }

    /**
     * Process sentiment data every 5 minutes for S&P 500 stocks
     */
    @Scheduled(fixedDelay = 300000, initialDelay = 30000) // 5 minutes, start after 30 seconds
    public void processSP500Sentiment() {
        log.info("======= Starting scheduled S&P 500 sentiment processing =======");

        // Reset counters
        successCount.set(0);
        failureCount.set(0);
        noDataCount.set(0);

        try {
            // Check connection first
            if (!nlpServiceClient.isConnected()) {
                log.error("NLP Service is not connected - skipping this run");
                return;
            }

            LocalDateTime endTime = LocalDateTime.now();
            LocalDateTime startTime = endTime.minusHours(24); // Last 24 hours for more data

            List<String> symbols = sp500Config.getSymbols();
            log.info("Processing sentiment for {} S&P 500 symbols from {} to {}",
                    symbols.size(), startTime, endTime);

            // Process first 10 symbols synchronously for testing
            List<String> testSymbols = symbols.stream().limit(10).collect(Collectors.toList());

            for (String symbol : testSymbols) {
                processSingleSymbol(symbol, startTime, endTime);
            }

            log.info("======= Completed S&P 500 sentiment processing =======");
            log.info("Results: Success={}, NoData={}, Failed={}",
                    successCount.get(), noDataCount.get(), failureCount.get());

        } catch (Exception e) {
            log.error("Error in scheduled sentiment processing: {}", e.getMessage(), e);
        }
    }

    private void processSingleSymbol(String symbol, LocalDateTime startTime, LocalDateTime endTime) {
        try {
            log.debug("Processing symbol: {}", symbol);

            // Get sentiment trends from NLP service
            List<SentimentTrendData> trends = nlpServiceClient.getSentimentTrends(
                    symbol, startTime, endTime, 24); // 24 hours of data

            if (trends == null || trends.isEmpty()) {
                log.warn("No sentiment data available for symbol: {}", symbol);
                noDataCount.incrementAndGet();
                return;
            }

            log.info("Found {} sentiment trends for {}", trends.size(), symbol);

            // Log first trend for debugging
            if (!trends.isEmpty()) {
                SentimentTrendData firstTrend = trends.get(0);
                log.debug("First trend for {}: sentiment={}, articles={}, volatility={}",
                        symbol, firstTrend.getAvgSentiment(),
                        firstTrend.getArticleCount(), firstTrend.getVolatility());
            }

            // Generate market impact prediction
            MarketPrediction prediction = predictionService.generatePredictionFromSentimentTrends(
                    symbol, trends);

            log.info("✓ Generated prediction for {}: direction={}, confidence={}, impact={}, change={}%",
                    symbol,
                    prediction.getDirection(),
                    prediction.getConfidence(),
                    prediction.getImpactScore(),
                    prediction.getPredictedChangePercent());

            // Notify SSE subscribers
            sp500Service.notifyPredictionUpdate(prediction);

            successCount.incrementAndGet();

        } catch (Exception e) {
            log.error("✗ Error processing symbol {}: {}", symbol, e.getMessage(), e);
            failureCount.incrementAndGet();
        }
    }

    /**
     * Cleanup old predictions (run daily at 2 AM)
     */
    @Scheduled(cron = "0 0 2 * * *")
    public void cleanupOldPredictions() {
        log.info("Starting cleanup of old predictions");
        try {
            sp500Service.cleanupOldPredictions(30); // Keep last 30 days
            log.info("Completed cleanup of old predictions");
        } catch (Exception e) {
            log.error("Error in cleanup: {}", e.getMessage(), e);
        }
    }
}