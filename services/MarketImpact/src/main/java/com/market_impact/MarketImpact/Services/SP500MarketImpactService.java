// services/MarketImpact/src/main/java/com/market_impact/MarketImpact/Services/SP500MarketImpactService.java
package com.market_impact.MarketImpact.Services;

import com.market_impact.MarketImpact.Repositories.MarketPredictionRepository;
import com.market_impact.MarketImpact.Config.SP500Config;
import com.market_impact.MarketImpact.dto.MarketImpactDto;
import com.market_impact.MarketImpact.dto.MarketSentimentSummary;
import com.market_impact.MarketImpact.entity.MarketPrediction;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.io.IOException;
import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
public class SP500MarketImpactService {

    private final MarketPredictionRepository predictionRepository;
    private final SP500Config sp500Config;

    // SSE emitters for real-time updates
    private final List<SseEmitter> emitters = new CopyOnWriteArrayList<>();

    /**
     * Get current market impact for all S&P 500 stocks
     */
    @Transactional(readOnly = true)
    public List<MarketImpactDto> getSP500MarketImpact() {
        log.info("Fetching market impact for S&P 500 stocks");

        List<String> symbols = sp500Config.getSymbols();

        return symbols.stream()
                .map(this::getLatestImpactForSymbol)
                .filter(Optional::isPresent)
                .map(Optional::get)
                .sorted(Comparator.comparing(MarketImpactDto::getImpactScore).reversed())
                .collect(Collectors.toList());
    }

    /**
     * Get market impact for specific symbols
     */
    @Transactional(readOnly = true)
    public List<MarketImpactDto> getMarketImpactForSymbols(List<String> symbols) {
        return symbols.stream()
                .map(this::getLatestImpactForSymbol)
                .filter(Optional::isPresent)
                .map(Optional::get)
                .collect(Collectors.toList());
    }

    /**
     * Get top movers (highest impact)
     */
    @Transactional(readOnly = true)
    public List<MarketImpactDto> getTopMovers(int limit) {
        List<MarketImpactDto> allImpacts = getSP500MarketImpact();
        return allImpacts.stream()
                .limit(limit)
                .collect(Collectors.toList());
    }

    /**
     * Get market impact by direction
     */
    @Transactional(readOnly = true)
    public Map<String, List<MarketImpactDto>> getMarketImpactByDirection() {
        List<MarketImpactDto> allImpacts = getSP500MarketImpact();

        return allImpacts.stream()
                .collect(Collectors.groupingBy(MarketImpactDto::getDirection));
    }

    /**
     * Get market sentiment summary
     */
    @Transactional(readOnly = true)
    public MarketSentimentSummary getMarketSentimentSummary() {
        List<MarketImpactDto> allImpacts = getSP500MarketImpact();

        long bullish = allImpacts.stream()
                .filter(i -> "UP".equals(i.getDirection()))
                .count();

        long bearish = allImpacts.stream()
                .filter(i -> "DOWN".equals(i.getDirection()))
                .count();

        long neutral = allImpacts.stream()
                .filter(i -> "NEUTRAL".equals(i.getDirection()))
                .count();

        double avgImpact = allImpacts.stream()
                .mapToDouble(i -> i.getImpactScore().doubleValue())
                .average()
                .orElse(0.0);

        double avgConfidence = allImpacts.stream()
                .mapToDouble(i -> i.getConfidence().doubleValue())
                .average()
                .orElse(0.0);

        return MarketSentimentSummary.builder()
                .timestamp(LocalDateTime.now())
                .totalStocks(allImpacts.size())
                .bullishCount(bullish)
                .bearishCount(bearish)
                .neutralCount(neutral)
                .bullishPercentage((double) bullish / allImpacts.size() * 100)
                .bearishPercentage((double) bearish / allImpacts.size() * 100)
                .averageImpactScore(BigDecimal.valueOf(avgImpact))
                .averageConfidence(BigDecimal.valueOf(avgConfidence))
                .marketSentiment(determineOverallMarketSentiment(bullish, bearish, neutral))
                .build();
    }

    /**
     * Register SSE emitter for real-time updates
     */
    public SseEmitter registerEmitter(Long timeout) {
        SseEmitter emitter = new SseEmitter(timeout);

        emitter.onCompletion(() -> {
            log.info("SSE connection completed");
            emitters.remove(emitter);
        });

        emitter.onTimeout(() -> {
            log.info("SSE connection timed out");
            emitters.remove(emitter);
        });

        emitter.onError((ex) -> {
            log.error("SSE connection error: {}", ex.getMessage());
            emitters.remove(emitter);
        });

        emitters.add(emitter);
        log.info("New SSE connection registered. Total connections: {}", emitters.size());

        return emitter;
    }

    /**
     * Notify all SSE subscribers of a prediction update
     */
    public void notifyPredictionUpdate(MarketPrediction prediction) {
        if (emitters.isEmpty()) {
            return;
        }

        MarketImpactDto impact = convertToDto(prediction);

        List<SseEmitter> deadEmitters = new ArrayList<>();

        for (SseEmitter emitter : emitters) {
            try {
                emitter.send(SseEmitter.event()
                        .name("prediction-update")
                        .data(impact));
            } catch (IOException e) {
                log.warn("Failed to send SSE event: {}", e.getMessage());
                deadEmitters.add(emitter);
            }
        }

        emitters.removeAll(deadEmitters);
    }

    /**
     * Cleanup old predictions
     */
    @Transactional
    public void cleanupOldPredictions(int daysToKeep) {
        LocalDateTime cutoffDate = LocalDateTime.now().minusDays(daysToKeep);

        // This would need a custom repository method
        log.info("Cleaning up predictions older than {}", cutoffDate);
        // predictionRepository.deleteByCreatedAtBefore(cutoffDate);
    }

    // Private helper methods

    private Optional<MarketImpactDto> getLatestImpactForSymbol(String symbol) {
        Optional<MarketPrediction> prediction =
                predictionRepository.findTopBySymbolOrderByPredictionTimestampDesc(symbol);

        return prediction.map(this::convertToDto);
    }

    private MarketImpactDto convertToDto(MarketPrediction prediction) {
        return MarketImpactDto.builder()
                .symbol(prediction.getSymbol())
                .predictedChangePercent(prediction.getPredictedChangePercent())
                .direction(prediction.getDirection())
                .confidence(prediction.getConfidence())
                .impactScore(prediction.getImpactScore())
                .modelType(prediction.getModelType())
                .timestamp(prediction.getPredictionTimestamp())
                .build();
    }

    private String determineOverallMarketSentiment(long bullish, long bearish, long neutral) {
        if (bullish > bearish * 1.5) {
            return "BULLISH";
        } else if (bearish > bullish * 1.5) {
            return "BEARISH";
        } else {
            return "NEUTRAL";
        }
    }
}