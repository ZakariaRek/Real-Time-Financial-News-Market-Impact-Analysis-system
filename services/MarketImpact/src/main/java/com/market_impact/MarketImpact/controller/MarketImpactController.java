package com.market_impact.MarketImpact.controller;

import com.market_impact.MarketImpact.Repositories.MarketPredictionRepository;
import com.market_impact.MarketImpact.Services.MarketImpactPredictionService;
import com.market_impact.MarketImpact.Services.SP500MarketImpactService;
import com.market_impact.MarketImpact.client.AlertSignalClient;
import com.market_impact.MarketImpact.client.NLPServiceClient;
import com.market_impact.MarketImpact.client.SentimentTrendData;
import com.market_impact.MarketImpact.entity.MarketPrediction;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/api/v1/market-impact")
@RequiredArgsConstructor
@Slf4j
public class MarketImpactController {

    private final MarketImpactPredictionService marketImpactService;
    private final NLPServiceClient nlpServiceClient;
    private final SP500MarketImpactService sp500Service;
    private final AlertSignalClient alertSignalClient;
    private final MarketPredictionRepository marketPredictionRepository; // Add this


    /**
     * Generate prediction - handles both articleId-based and trend-based predictions
     */
    @PostMapping("/predict")
    public ResponseEntity<?> generatePrediction(@RequestBody PredictionRequest request) {

        log.info("REST: Generate prediction for article: {} and symbol: {}",
                request.getArticleId(), request.getSymbol());

        try {
            MarketPrediction prediction;

            // If articleId is provided, use article-based prediction
            if (request.getArticleId() != null) {
                prediction = marketImpactService.generatePredictionFromSentiment(
                        request.getArticleId(),
                        request.getSymbol());
            } else {
                // Otherwise, use trend-based prediction
                log.info("No articleId provided, using trend-based prediction");

                LocalDateTime endTime = LocalDateTime.now();
                LocalDateTime startTime = endTime.minusHours(24);

                List<SentimentTrendData> trends = nlpServiceClient.getSentimentTrends(
                        request.getSymbol(), startTime, endTime, 24);

                if (trends == null || trends.isEmpty()) {
                    return ResponseEntity.status(HttpStatus.NOT_FOUND)
                            .body("No sentiment data found for " + request.getSymbol());
                }

                prediction = marketImpactService.generatePredictionFromSentimentTrends(
                        request.getSymbol(), trends);
            }

            // ✅ Notify SSE subscribers
            sp500Service.notifyPredictionUpdate(prediction);

            // ✅ Notify AlertSignal via gRPC
            alertSignalClient.notifyPrediction(
                    prediction.getId().toString(),
                    prediction.getSymbol(),
                    prediction.getConfidence().doubleValue(),
                    prediction.getImpactScore().doubleValue()
            );

            log.info("✅ Prediction created and notifications sent for {}", request.getSymbol());
            return ResponseEntity.ok(prediction);

        } catch (Exception e) {
            log.error("Failed to generate prediction: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body("Error: " + e.getMessage());
        }
    }

    /**
     * Generate prediction from trends only (no article required)
     */
    @PostMapping("/predict/trends")
    public ResponseEntity<?> generatePredictionFromTrends(
            @RequestBody TrendPredictionRequest request) {

        log.info("REST: Generate prediction from trends for symbol: {}", request.getSymbol());

        try {
            LocalDateTime endTime = LocalDateTime.now();
            LocalDateTime startTime = request.getHoursBack() != null
                    ? endTime.minusHours(request.getHoursBack())
                    : endTime.minusHours(24);

            List<SentimentTrendData> trends = nlpServiceClient.getSentimentTrends(
                    request.getSymbol(), startTime, endTime, 24);

            if (trends == null || trends.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                        .body("No sentiment data found for " + request.getSymbol());
            }

            MarketPrediction prediction = marketImpactService.generatePredictionFromSentimentTrends(
                    request.getSymbol(), trends);

            // ✅ Notify SSE subscribers
            sp500Service.notifyPredictionUpdate(prediction);

            // ✅ Notify AlertSignal via gRPC
            alertSignalClient.notifyPrediction(
                    prediction.getId().toString(),
                    prediction.getSymbol(),
                    prediction.getConfidence().doubleValue(),
                    prediction.getImpactScore().doubleValue()
            );

            return ResponseEntity.ok(prediction);
        } catch (Exception e) {
            log.error("Failed to generate prediction from trends: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body("Error: " + e.getMessage());
        }
    }

    /**
     * TEST: Create prediction with mock data
     */
    @PostMapping("/predict/test")
    public ResponseEntity<?> generateTestPrediction(
            @RequestParam(defaultValue = "AAPL") String symbol) {

        log.info("TEST: Creating test prediction for {}", symbol);

        try {
            // Create mock sentiment trends
            List<SentimentTrendData> mockTrends = new java.util.ArrayList<>();
            LocalDateTime now = LocalDateTime.now();

            for (int i = 0; i < 24; i++) {
                SentimentTrendData trend = SentimentTrendData.builder()
                        .timestamp(now.minusHours(i))
                        .symbol(symbol)
                        .avgSentiment(0.5 + (Math.random() * 0.3))
                        .articleCount(5 + (long)(Math.random() * 10))
                        .volatility(0.1 + (Math.random() * 0.2))
                        .build();
                mockTrends.add(trend);
            }

            MarketPrediction prediction = marketImpactService.generatePredictionFromSentimentTrends(
                    symbol, mockTrends);

            // ✅ Notify SSE subscribers
            sp500Service.notifyPredictionUpdate(prediction);

            // ✅ Notify AlertSignal via gRPC
            alertSignalClient.notifyPrediction(
                    prediction.getId().toString(),
                    prediction.getSymbol(),
                    prediction.getConfidence().doubleValue(),
                    prediction.getImpactScore().doubleValue()
            );

            log.info("✅ Test prediction created and sent to AlertSignal for {}", symbol);
            return ResponseEntity.ok(prediction);

        } catch (Exception e) {
            log.error("❌ Failed to create test prediction: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body("Error: " + e.getMessage());
        }
    }

    @PostMapping("/predict/batch")
    public ResponseEntity<List<MarketPrediction>> batchGeneratePredictions(
            @RequestBody BatchPredictionRequest request) {

        log.info("REST: Batch generate predictions for {} articles", request.getArticleIds().size());

        try {
            List<MarketPrediction> predictions = marketImpactService.batchGeneratePredictions(
                    request.getArticleIds(),
                    request.getSymbol());

            // Notify for each prediction
            predictions.forEach(prediction -> {
                sp500Service.notifyPredictionUpdate(prediction);
                alertSignalClient.notifyPrediction(
                        prediction.getId().toString(),
                        prediction.getSymbol(),
                        prediction.getConfidence().doubleValue(),
                        prediction.getImpactScore().doubleValue()
                );
            });

            return ResponseEntity.ok(predictions);
        } catch (Exception e) {
            log.error("Failed to batch generate predictions: {}", e.getMessage());
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    /**
     * TEST: Create prediction with HIGH confidence and impact (will trigger AlertSignal)
     */
    @PostMapping("/predict/test-high")
    public ResponseEntity<?> generateHighConfidencePrediction(
            @RequestParam(defaultValue = "AAPL") String symbol,
            @RequestParam(defaultValue = "0.85") double confidence,
            @RequestParam(defaultValue = "75.0") double impactScore) {

        log.info("TEST: Creating HIGH confidence prediction for {} (confidence={}, impact={})",
                symbol, confidence, impactScore);

        try {
            // Create mock sentiment trends with high values
            List<SentimentTrendData> mockTrends = new java.util.ArrayList<>();
            LocalDateTime now = LocalDateTime.now();

            // Create very positive sentiment trends
            for (int i = 0; i < 24; i++) {
                SentimentTrendData trend = SentimentTrendData.builder()
                        .timestamp(now.minusHours(i))
                        .symbol(symbol)
                        .avgSentiment(0.7 + (Math.random() * 0.2)) // High positive sentiment
                        .articleCount(15 + (long)(Math.random() * 20)) // Many articles
                        .volatility(0.15 + (Math.random() * 0.1))
                        .build();
                mockTrends.add(trend);
            }

            MarketPrediction prediction = marketImpactService.generatePredictionFromSentimentTrends(
                    symbol, mockTrends);

            // Override with guaranteed high values if needed
            if (prediction.getConfidence().doubleValue() < 0.6) {
                prediction.setConfidence(new java.math.BigDecimal(confidence));
            }
            if (prediction.getImpactScore().doubleValue() < 40.0) {
                prediction.setImpactScore(new java.math.BigDecimal(impactScore));
            }

            // Save the updated prediction
            prediction = marketPredictionRepository.save(prediction);

            // ✅ Notify SSE subscribers
            sp500Service.notifyPredictionUpdate(prediction);

            // ✅ Notify AlertSignal via gRPC
            alertSignalClient.notifyPrediction(
                    prediction.getId().toString(),
                    prediction.getSymbol(),
                    prediction.getConfidence().doubleValue(),
                    prediction.getImpactScore().doubleValue()
            );

            log.info("✅ HIGH confidence test prediction created: confidence={}, impact={}",
                    prediction.getConfidence(), prediction.getImpactScore());
            return ResponseEntity.ok(prediction);

        } catch (Exception e) {
            log.error("❌ Failed to create high confidence test prediction: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body("Error: " + e.getMessage());
        }
    }

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class PredictionRequest {
        private UUID articleId; // Now optional
        private String symbol;
    }

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class TrendPredictionRequest {
        private String symbol;
        private Integer hoursBack;
    }

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class BatchPredictionRequest {
        private List<UUID> articleIds;
        private String symbol;
    }
}