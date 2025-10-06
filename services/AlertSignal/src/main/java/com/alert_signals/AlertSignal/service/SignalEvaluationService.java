package com.alert_signals.AlertSignal.service;

import com.alert_signals.AlertSignal.client.MarketImpactClient;
import com.alert_signals.AlertSignal.dto.grpc.MarketPredictionDto;
import com.alert_signals.AlertSignal.dto.grpc.RiskMetricsDto;
import com.alert_signals.AlertSignal.entity.SignalRules;
import com.alert_signals.AlertSignal.entity.TradingSignal;
import com.alert_signals.AlertSignal.repository.SignalRulesRepository;
import com.alert_signals.AlertSignal.repository.TradingSignalRepository;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.RoundingMode;
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
    private static final BigDecimal HIGH_CONFIDENCE_THRESHOLD = new BigDecimal("0.8");
    private static final BigDecimal MIN_IMPACT_SCORE = new BigDecimal("50.0");
    private static final String ACTIVE_SIGNALS_KEY = "active_signals:";
    private static final long SIGNAL_TTL_HOURS = 24;

    /**
     * Process prediction from MarketImpact service
     */
    public TradingSignal processPrediction(UUID predictionId) {
        log.info("Processing prediction: {}", predictionId);

        try {
            // 1. Get prediction from MarketImpact service (now returns DTO)
            MarketPredictionDto prediction = marketImpactClient.getPrediction(predictionId);

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

            // 5. Check confidence threshold (now using BigDecimal comparison)
            if (prediction.getConfidence().compareTo(HIGH_CONFIDENCE_THRESHOLD) < 0) {
                log.info("Low confidence signal discarded: {} (confidence: {})",
                        predictionId, prediction.getConfidence());
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
    private boolean evaluateSignalRules(MarketPredictionDto prediction) {
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
    private boolean evaluateRule(SignalRules rule, MarketPredictionDto prediction) {
        try {
            // Parse rule conditions (stored as JSON)
            RuleConditions conditions = objectMapper.readValue(
                    rule.getConditions(),
                    RuleConditions.class
            );

            // Direct BigDecimal comparison - no parsing needed!
            BigDecimal confidence = prediction.getConfidence();
            BigDecimal impactScore = prediction.getImpactScore();

            // **ADD THIS LOGGING**
            log.info("Evaluating rule '{}' for {}: confidence={} (min={}), impact={} (min={}), direction={} (required={})",
                    rule.getRuleName(), prediction.getSymbol(),
                    confidence, conditions.minConfidence,
                    impactScore, conditions.minImpactScore,
                    prediction.getDirection(), conditions.requiredDirection);

            // Check conditions
            if (conditions.minConfidence != null &&
                    confidence.compareTo(conditions.minConfidence) < 0) {
                log.info("Rule '{}' FAILED: confidence {} < {}",
                        rule.getRuleName(), confidence, conditions.minConfidence);
                return false;
            }

            if (conditions.minImpactScore != null &&
                    impactScore.compareTo(conditions.minImpactScore) < 0) {
                log.info("Rule '{}' FAILED: impactScore {} < {}",
                        rule.getRuleName(), impactScore, conditions.minImpactScore);
                return false;
            }

            if (conditions.requiredDirection != null &&
                    !prediction.getDirection().equals(conditions.requiredDirection)) {
                log.info("Rule '{}' FAILED: direction {} != {}",
                        rule.getRuleName(), prediction.getDirection(), conditions.requiredDirection);
                return false;
            }

            log.info("Rule '{}' PASSED!", rule.getRuleName());
            return true;

        } catch (Exception e) {
            log.error("Error evaluating rule {}: {}", rule.getRuleName(), e.getMessage());
            return false;
        }
    }
    /**
     * Apply risk filters to prediction
     */
    private boolean applyRiskFilters(MarketPredictionDto prediction) {
        if (prediction.getRiskMetrics() == null || prediction.getRiskMetrics().isEmpty()) {
            log.warn("No risk metrics available for prediction: {}", prediction.getId());
            return true; // Allow if no risk metrics
        }

        RiskMetricsDto risk = prediction.getRiskMetrics().get(0); // Assume first metric

        // Check risk level
        if ("CRITICAL".equals(risk.getRiskLevel())) {
            log.warn("Critical risk level detected for symbol: {}", prediction.getSymbol());
            return false;
        }

        // Check VaR threshold (now using BigDecimal)
        BigDecimal maxVaR = new BigDecimal("10.0");
        if (risk.getVar951Day().compareTo(maxVaR) > 0) {
            log.warn("VaR too high: {} for symbol: {}",
                    risk.getVar951Day(), prediction.getSymbol());
            return false;
        }

        return true;
    }

    /**
     * Calculate signal strength based on prediction metrics
     */
    private BigDecimal calculateSignalStrength(MarketPredictionDto prediction) {
        BigDecimal confidence = prediction.getConfidence();
        BigDecimal impactScore = prediction.getImpactScore();
        BigDecimal predictedChange = prediction.getPredictedChangePercent().abs();

        // Weighted combination using BigDecimal
        BigDecimal strengthPart1 = confidence.multiply(new BigDecimal("0.4"));
        BigDecimal strengthPart2 = impactScore.divide(new BigDecimal("100"), 4, RoundingMode.HALF_UP)
                .multiply(new BigDecimal("0.4"));

        // Min of (predictedChange / 10, 1.0) * 0.2
        BigDecimal changeFactor = predictedChange.divide(new BigDecimal("10"), 4, RoundingMode.HALF_UP);
        if (changeFactor.compareTo(BigDecimal.ONE) > 0) {
            changeFactor = BigDecimal.ONE;
        }
        BigDecimal strengthPart3 = changeFactor.multiply(new BigDecimal("0.2"));

        BigDecimal strength = strengthPart1.add(strengthPart2).add(strengthPart3);
        return strength.setScale(4, RoundingMode.HALF_UP);
    }

    /**
     * Create trading signal from prediction
     */
    private TradingSignal createTradingSignal(
            MarketPredictionDto prediction,
            BigDecimal strength) {

        return TradingSignal.builder()
                .predictionId(prediction.getId())
                .symbol(prediction.getSymbol())
                .signalType("PREDICTION_BASED")
                .direction(prediction.getDirection())
                .strength(strength)
                .confidence(prediction.getConfidence())
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
        public BigDecimal minConfidence;
        public BigDecimal minImpactScore;
        public String requiredDirection;
    }
}