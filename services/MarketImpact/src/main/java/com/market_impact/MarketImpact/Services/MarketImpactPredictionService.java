package com.market_impact.MarketImpact.Services;

import com.market_impact.MarketImpact.Repositories.MarketPredictionRepository;
import com.market_impact.MarketImpact.client.NLPServiceClient;
import com.market_impact.MarketImpact.client.SentimentData;
import com.market_impact.MarketImpact.client.SentimentTrendData;
import com.market_impact.MarketImpact.entity.MarketPrediction;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
@Transactional
public class MarketImpactPredictionService {

    private final NLPServiceClient nlpServiceClient;
    private final MarketPredictionRepository marketPredictionRepository;

    // Sentiment threshold constants
    private static final double STRONG_POSITIVE_THRESHOLD = 0.5;
    private static final double POSITIVE_THRESHOLD = 0.2;
    private static final double NEGATIVE_THRESHOLD = -0.2;
    private static final double STRONG_NEGATIVE_THRESHOLD = -0.5;

    // Impact score multipliers
    private static final double BASE_IMPACT_MULTIPLIER = 10.0;
    private static final double CONFIDENCE_WEIGHT = 0.7;
    private static final double SENTIMENT_WEIGHT = 0.3;

    /**
     * Generate market prediction from sentiment analysis
     */
    public MarketPrediction generatePredictionFromSentiment(UUID articleId, String symbol) {
        log.info("Generating market prediction for article: {} and symbol: {}", articleId, symbol);

        // Get sentiment data from NLP service
        SentimentData sentiment = nlpServiceClient.getSentimentForArticle(articleId.toString())
                .orElseThrow(() -> new RuntimeException("Sentiment data not found for article: " + articleId));

        // Get recent sentiment trends for context
        LocalDateTime endTime = LocalDateTime.now();
        LocalDateTime startTime = endTime.minusHours(24);
        List<SentimentTrendData> trends = nlpServiceClient.getSentimentTrends(
                symbol, startTime, endTime, 24);

        // Calculate prediction metrics
        BigDecimal predictedChange = calculatePredictedChange(sentiment, trends);
        String direction = determineDirection(predictedChange);
        BigDecimal confidence = calculateConfidence(sentiment, trends);
        BigDecimal impactScore = calculateImpactScore(sentiment, trends);

        // Create and save prediction
        MarketPrediction prediction = MarketPrediction.builder()
                .articleId(articleId)
                .symbol(symbol)
                .predictedChangePercent(predictedChange)
                .direction(direction)
                .confidence(confidence)
                .impactScore(impactScore)
                .modelType("SENTIMENT_BASED_v1.0")
                .predictionTimestamp(LocalDateTime.now())
                .build();

        MarketPrediction savedPrediction = marketPredictionRepository.save(prediction);
        log.info("Market prediction saved: {}", savedPrediction.getId());

        return savedPrediction;
    }

    /**
     * Calculate predicted percentage change based on sentiment
     */
    private BigDecimal calculatePredictedChange(SentimentData sentiment, List<SentimentTrendData> trends) {
        double sentimentScore = sentiment.getCompoundScore();
        double confidence = sentiment.getConfidence();

        // Base prediction from sentiment
        double basePrediction = sentimentScore * BASE_IMPACT_MULTIPLIER;

        // Adjust based on sentiment trends
        double trendAdjustment = calculateTrendAdjustment(trends);

        // Adjust based on confidence
        double confidenceAdjustment = confidence * 0.5;

        double totalPrediction = basePrediction * (1 + trendAdjustment + confidenceAdjustment);

        // Cap the prediction at realistic bounds (-10% to +10%)
        totalPrediction = Math.max(-10.0, Math.min(10.0, totalPrediction));

        return BigDecimal.valueOf(totalPrediction).setScale(4, RoundingMode.HALF_UP);
    }

    /**
     * Calculate trend adjustment factor
     */
    private double calculateTrendAdjustment(List<SentimentTrendData> trends) {
        if (trends == null || trends.isEmpty()) {
            return 0.0;
        }

        // Calculate sentiment momentum
        double recentAvg = trends.stream()
                .limit(6) // Last 6 hours
                .mapToDouble(SentimentTrendData::getAvgSentiment)
                .average()
                .orElse(0.0);

        double olderAvg = trends.stream()
                .skip(6)
                .mapToDouble(SentimentTrendData::getAvgSentiment)
                .average()
                .orElse(0.0);

        double momentum = recentAvg - olderAvg;

        // Calculate volatility factor
        double avgVolatility = trends.stream()
                .mapToDouble(SentimentTrendData::getVolatility)
                .average()
                .orElse(0.0);

        // Higher volatility increases potential impact
        double volatilityFactor = avgVolatility * 0.2;

        return momentum + volatilityFactor;
    }

    /**
     * Determine market direction
     */
    private String determineDirection(BigDecimal predictedChange) {
        double change = predictedChange.doubleValue();

        if (change > POSITIVE_THRESHOLD) {
            return "UP";
        } else if (change < NEGATIVE_THRESHOLD) {
            return "DOWN";
        } else {
            return "NEUTRAL";
        }
    }

    /**
     * Calculate prediction confidence
     */
    private BigDecimal calculateConfidence(SentimentData sentiment, List<SentimentTrendData> trends) {
        double sentimentConfidence = sentiment.getConfidence();

        // Higher confidence if sentiment is strong
        double sentimentStrength = Math.abs(sentiment.getCompoundScore());

        // Higher confidence if trends are consistent
        double trendConsistency = calculateTrendConsistency(trends);

        // Weighted average
        double totalConfidence = (sentimentConfidence * CONFIDENCE_WEIGHT) +
                (sentimentStrength * SENTIMENT_WEIGHT) +
                (trendConsistency * 0.3);

        // Normalize to 0-1 range
        totalConfidence = Math.max(0.0, Math.min(1.0, totalConfidence));

        return BigDecimal.valueOf(totalConfidence).setScale(4, RoundingMode.HALF_UP);
    }

    /**
     * Calculate trend consistency (how aligned recent trends are)
     */
    private double calculateTrendConsistency(List<SentimentTrendData> trends) {
        if (trends == null || trends.size() < 2) {
            return 0.5; // Default medium consistency
        }

        // Calculate standard deviation of sentiment
        double mean = trends.stream()
                .mapToDouble(SentimentTrendData::getAvgSentiment)
                .average()
                .orElse(0.0);

        double variance = trends.stream()
                .mapToDouble(t -> Math.pow(t.getAvgSentiment() - mean, 2))
                .average()
                .orElse(0.0);

        double stdDev = Math.sqrt(variance);

        // Lower standard deviation = higher consistency
        // Normalize: std dev of 0 = consistency of 1, std dev of 1+ = consistency of 0
        return Math.max(0.0, 1.0 - stdDev);
    }

    /**
     * Calculate overall impact score
     */
    private BigDecimal calculateImpactScore(SentimentData sentiment, List<SentimentTrendData> trends) {
        double sentimentMagnitude = Math.abs(sentiment.getCompoundScore());
        double confidence = sentiment.getConfidence();

        // Article volume factor
        long totalArticles = trends.stream()
                .mapToLong(SentimentTrendData::getArticleCount)
                .sum();
        double volumeFactor = Math.min(1.0, totalArticles / 50.0); // Normalize to max of 50 articles

        // Calculate impact
        double impact = (sentimentMagnitude * confidence * 100) * (1 + volumeFactor);

        // Cap at 100
        impact = Math.min(100.0, impact);

        return BigDecimal.valueOf(impact).setScale(4, RoundingMode.HALF_UP);
    }

    /**
     * Batch process predictions for multiple articles
     */
    public List<MarketPrediction> batchGeneratePredictions(List<UUID> articleIds, String symbol) {
        log.info("Batch generating predictions for {} articles", articleIds.size());

        return articleIds.stream()
                .map(articleId -> {
                    try {
                        return generatePredictionFromSentiment(articleId, symbol);
                    } catch (Exception e) {
                        log.error("Failed to generate prediction for article {}: {}", articleId, e.getMessage());
                        return null;
                    }
                })
                .filter(prediction -> prediction != null)
                .toList();
    }
}