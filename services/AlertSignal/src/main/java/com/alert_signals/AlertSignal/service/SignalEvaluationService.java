package com.alert_signals.AlertSignal.service;

import com.alert_signals.AlertSignal.client.MarketImpactClient;
import com.alert_signals.AlertSignal.entity.SignalRules;
import com.alert_signals.AlertSignal.entity.TradingSignal;
import com.alert_signals.AlertSignal.repository.SignalRulesRepository;
import com.alert_signals.AlertSignal.repository.TradingSignalRepository;
import com.alert_signals.AlertSignal.grpc.generated.MarketPrediction;
import com.alert_signals.AlertSignal.grpc.generated.RiskMetrics;

import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;
import java.util.concurrent.TimeUnit;

@Service
@RequiredArgsConstructor
@Slf4j
@Transactional
public class SignalEvaluationService {

    private final MarketImpactClient marketImpactClient;
    private final SignalRulesRepository signalRulesRepository;
    private final TradingSignalRepository tradingSignalRepository;
    private final RedisTemplate<String, Object> redisTemplate;
    private final ObjectMapper objectMapper;

    // Constants
    private static final double HIGH_CONFIDENCE_THRESHOLD = 0.8;
    private static final double MIN_IMPACT_SCORE = 50.0;
    private static final String ACTIVE_SIGNALS_KEY = "active_signals:";
    private static final long SIGNAL_TTL_HOURS = 24;

    /**
     * Process prediction from MarketImpact service
     * This is the main entry point called via gRPC
     */
    public TradingSignal processPrediction(UUID predictionId) {
        log.info("Processing prediction: {}", predictionId);

        try {
            // 1. Get prediction from MarketImpact service
            MarketPrediction prediction = marketImpactClient.getPrediction(predictionId);

            // 2. Evaluate signal rules
            boolean rulesPass = evaluateSignalRules(prediction);
            if (!rulesPass) {
                log.info("Signal rules not met for prediction: {}", predictionId);
                return null;
            }

            // 3. Apply risk filters
            boolean riskPass = applyRiskFilters(prediction);
            if (!riskPass) {
                log.info("Risk filters not passed for prediction: {}", predictionId);
                return null;
            }

            // 4. Calculate signal strength
            BigDecimal signalStrength = calculateSignalStrength(prediction);

            // 5. Check confidence threshold
            double confidence = Double.parseDouble(prediction.getConfidence().getValue());
            if (confidence < HIGH_CONFIDENCE_THRESHOLD) {
                log.info("Low confidence signal discarded: {} (confidence: {})",
                        predictionId, confidence);
                return null;
            }

            // 6. Create and store trading signal
            TradingSignal signal = createTradingSignal(prediction, signalStrength);
            TradingSignal savedSignal = tradingSignalRepository.save(signal);

            // 7. Cache active signal in Redis
            cacheActiveSignal(savedSignal);

            log.info("High-confidence signal created: {} for symbol: {}",
                    savedSignal.getId(), savedSignal.getSymbol());

            return savedSignal;

        } catch (Exception e) {
            log.error("Error processing prediction {}: {}", predictionId, e.getMessage(), e);
            return null;
        }
    }

    /**
     * Evaluate signal rules against prediction
     */
    private boolean evaluateSignalRules(MarketPrediction prediction) {
        String symbol = prediction.getSymbol();

        // Get active rules for this symbol
        List<SignalRules> activeRules = signalRulesRepository
                .findActiveRulesBySymbol(symbol);

        if (activeRules.isEmpty()) {
            log.warn("No active rules found for symbol: {}", symbol);
            return false;
        }

        // Check if prediction meets any active rule criteria
        for (SignalRules rule : activeRules) {
            if (evaluateRule(rule, prediction)) {
                log.info("Signal rule '{}' passed for symbol: {}",
                        rule.getRuleName(), symbol);
                return true;
            }
        }

        return false;
    }

    /**
     * Evaluate a single rule against prediction
     */
    private boolean evaluateRule(SignalRules rule, MarketPrediction prediction) {
        try {
            // Parse rule conditions (stored as JSON)
            RuleConditions conditions = objectMapper.readValue(
                    rule.getConditions(),
                    RuleConditions.class
            );

            double confidence = Double.parseDouble(
                    prediction.getConfidence().getValue()
            );
            double impactScore = Double.parseDouble(
                    prediction.getImpactScore().getValue()
            );

            // Check conditions
            if (conditions.minConfidence != null &&
                    confidence < conditions.minConfidence) {
                return false;
            }

            if (conditions.minImpactScore != null &&
                    impactScore < conditions.minImpactScore) {
                return false;
            }

            if (conditions.requiredDirection != null &&
                    !prediction.getDirection().name().equals(conditions.requiredDirection)) {
                return false;
            }

            return true;

        } catch (Exception e) {
            log.error("Error evaluating rule {}: {}", rule.getRuleName(), e.getMessage());
            return false;
        }
    }

    /**
     * Apply risk filters to prediction
     */
    private boolean applyRiskFilters(MarketPrediction prediction) {
        if (prediction.getRiskMetricsCount() == 0) {
            log.warn("No risk metrics available for prediction: {}",
                    prediction.getId().getValue());
            return true; // Allow if no risk metrics
        }

        RiskMetrics risk = prediction.getRiskMetrics(0);

        // Check risk level
        if (risk.getRiskLevel().name().equals("CRITICAL")) {
            log.warn("Critical risk level detected for symbol: {}",
                    prediction.getSymbol());
            return false;
        }

        // Check VaR threshold
        double var = Double.parseDouble(risk.getVar951Day().getValue());
        if (var > 10.0) { // 10% max VaR
            log.warn("VaR too high: {} for symbol: {}", var, prediction.getSymbol());
            return false;
        }

        return true;
    }

    /**
     * Calculate signal strength based on prediction metrics
     */
    private BigDecimal calculateSignalStrength(MarketPrediction prediction) {
        double confidence = Double.parseDouble(prediction.getConfidence().getValue());
        double impactScore = Double.parseDouble(prediction.getImpactScore().getValue());
        double predictedChange = Math.abs(Double.parseDouble(
                prediction.getPredictedChangePercent().getValue()
        ));

        // Weighted combination
        double strength = (confidence * 0.4) +
                (impactScore / 100.0 * 0.4) +
                (Math.min(predictedChange / 10.0, 1.0) * 0.2);

        return BigDecimal.valueOf(strength).setScale(4, RoundingMode.HALF_UP);
    }

    /**
     * Create trading signal from prediction
     */
    private TradingSignal createTradingSignal(
            MarketPrediction prediction,
            BigDecimal strength) {

        return TradingSignal.builder()
                .predictionId(UUID.fromString(prediction.getId().getValue()))
                .symbol(prediction.getSymbol())
                .signalType("PREDICTION_BASED")
                .direction(prediction.getDirection().name())
                .strength(strength)
                .confidence(new BigDecimal(prediction.getConfidence().getValue()))
                .status("ACTIVE")
                .build();
    }

    /**
     * Cache active signal in Redis
     */
    private void cacheActiveSignal(TradingSignal signal) {
        try {
            String key = ACTIVE_SIGNALS_KEY + signal.getSymbol();
            redisTemplate.opsForValue().set(
                    key,
                    signal,
                    SIGNAL_TTL_HOURS,
                    TimeUnit.HOURS
            );
            log.debug("Cached active signal in Redis: {}", key);
        } catch (Exception e) {
            log.error("Failed to cache signal in Redis: {}", e.getMessage());
        }
    }

    /**
     * Get active signals from Redis cache
     */
    public List<TradingSignal> getActiveSignalsFromCache(String symbol) {
        try {
            String key = ACTIVE_SIGNALS_KEY + symbol;
            Object cached = redisTemplate.opsForValue().get(key);
            if (cached != null) {
                return List.of((TradingSignal) cached);
            }
        } catch (Exception e) {
            log.error("Failed to get cached signals: {}", e.getMessage());
        }
        return List.of();
    }

    // Inner class for rule conditions
    private static class RuleConditions {
        public Double minConfidence;
        public Double minImpactScore;
        public String requiredDirection;
    }
}