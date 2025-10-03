package com.alert_signals.AlertSignal.service;

import com.alert_signals.AlertSignal.controller.SignalWebSocketController;
import com.alert_signals.AlertSignal.entity.AlertDeliveryLog;
import com.alert_signals.AlertSignal.entity.SignalPerformance;
import com.alert_signals.AlertSignal.entity.TradingSignal;
import com.alert_signals.AlertSignal.entity.UserSubscription;
import com.alert_signals.AlertSignal.repository.SignalPerformanceRepository;
import com.alert_signals.AlertSignal.repository.UserSubscriptionRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDate;
import java.util.List;

/**
 * Service to handle signal notifications and performance updates
 * Implements the signal performance update flow from the diagram
 */
@Service
@RequiredArgsConstructor
@Slf4j
@Transactional
public class SignalNotificationService {

    private final SignalWebSocketController webSocketController;
    private final SignalPerformanceRepository signalPerformanceRepository;
    private final UserSubscriptionRepository userSubscriptionRepository;
    private final AlertDeliveryService alertDeliveryService;

    /**
     * Process new signal and notify subscribers
     * Called after a high-confidence signal is created
     */
    public void processNewSignal(TradingSignal signal) {
        log.info("Processing new signal: {} for symbol: {}",
                signal.getId(), signal.getSymbol());

        try {
            // 1. Broadcast signal via WebSocket
            webSocketController.broadcastNewSignal(signal);

            // 2. Find matching subscriptions
            List<UserSubscription> subscriptions = findMatchingSubscriptions(signal);

            // 3. Deliver alerts to subscribed users
            for (UserSubscription subscription : subscriptions) {
                deliverAlert(signal, subscription);
            }

            // 4. Initialize signal performance tracking
            initializePerformanceTracking(signal);

            log.info("Successfully processed signal: {}", signal.getId());

        } catch (Exception e) {
            log.error("Error processing new signal: {}", e.getMessage(), e);
        }
    }

    /**
     * Find user subscriptions that match the signal criteria
     */
    private List<UserSubscription> findMatchingSubscriptions(TradingSignal signal) {
        // Get subscriptions for this symbol
        List<UserSubscription> symbolSubscriptions =
                userSubscriptionRepository.findActiveSubscriptionsBySymbol(signal.getSymbol());

        // Filter by signal type and confidence threshold
        return symbolSubscriptions.stream()
                .filter(sub -> matchesSubscription(signal, sub))
                .toList();
    }

    /**
     * Check if signal matches subscription criteria
     */
    private boolean matchesSubscription(TradingSignal signal, UserSubscription subscription) {
        // Check signal type
        if (subscription.getSignalTypes() != null &&
                !subscription.getSignalTypes().contains(signal.getSignalType())) {
            return false;
        }

        // Check minimum confidence
        if (subscription.getMinConfidence() != null &&
                signal.getConfidence().compareTo(subscription.getMinConfidence()) < 0) {
            return false;
        }

        return true;
    }

    /**
     * Deliver alert to user via configured delivery methods
     */
    private void deliverAlert(TradingSignal signal, UserSubscription subscription) {
        if (subscription.getDeliveryMethods() == null) {
            return;
        }

        for (String method : subscription.getDeliveryMethods()) {
            try {
                long startTime = System.currentTimeMillis();

                // Deliver via specified method (email, SMS, push, etc.)
                deliverViaMethod(signal, subscription, method);

                int latency = (int) (System.currentTimeMillis() - startTime);

                // Log delivery
                alertDeliveryService.logDelivery(
                        AlertDeliveryLog.builder()
                                .signalId(signal.getId())
                                .subscriptionId(subscription.getId())
                                .deliveryMethod(method)
                                .status("SUCCESS")
                                .deliveryLatencyMs(latency)
                                .build()
                );

                log.info("Alert delivered via {} to subscription: {}",
                        method, subscription.getId());

            } catch (Exception e) {
                log.error("Failed to deliver alert via {}: {}", method, e.getMessage());

                // Log failed delivery
                alertDeliveryService.logDelivery(
                        AlertDeliveryLog.builder()
                                .signalId(signal.getId())
                                .subscriptionId(subscription.getId())
                                .deliveryMethod(method)
                                .status("FAILED")
                                .build()
                );
            }
        }
    }

    /**
     * Deliver alert via specific method
     * TODO: Implement actual delivery mechanisms
     */
    private void deliverViaMethod(
            TradingSignal signal,
            UserSubscription subscription,
            String method) {

        switch (method.toUpperCase()) {
            case "EMAIL":
                // TODO: Send email
                log.debug("Would send email for signal: {}", signal.getId());
                break;
            case "SMS":
                // TODO: Send SMS
                log.debug("Would send SMS for signal: {}", signal.getId());
                break;
            case "PUSH":
                // TODO: Send push notification
                log.debug("Would send push notification for signal: {}", signal.getId());
                break;
            case "WEBHOOK":
                // TODO: Call webhook
                log.debug("Would call webhook for signal: {}", signal.getId());
                break;
            default:
                log.warn("Unknown delivery method: {}", method);
        }
    }

    /**
     * Initialize performance tracking for new signal
     * Implements: "ASS->>+PGSIGNAL: Update signal performance"
     */
    private void initializePerformanceTracking(TradingSignal signal) {
        try {
            SignalPerformance performance = SignalPerformance.builder()
                    .signalId(signal.getId())
                    .performanceDate(LocalDate.now())
                    .return1d(BigDecimal.ZERO)
                    .return1w(BigDecimal.ZERO)
                    .maxDrawdown(BigDecimal.ZERO)
                    .accuracy(BigDecimal.ZERO)
                    .build();

            signalPerformanceRepository.save(performance);
            log.info("Initialized performance tracking for signal: {}", signal.getId());

        } catch (Exception e) {
            log.error("Failed to initialize performance tracking: {}", e.getMessage());
        }
    }

    /**
     * Update signal performance metrics
     * Called periodically to update actual returns
     */
    public void updateSignalPerformance(
            TradingSignal signal,
            BigDecimal actualReturn1d,
            BigDecimal actualReturn1w,
            BigDecimal maxDrawdown) {

        try {
            SignalPerformance performance = signalPerformanceRepository
                    .findBySignalIdAndPerformanceDate(signal.getId(), LocalDate.now())
                    .orElseGet(() -> SignalPerformance.builder()
                            .signalId(signal.getId())
                            .performanceDate(LocalDate.now())
                            .build());

            performance.setReturn1d(actualReturn1d);
            performance.setReturn1w(actualReturn1w);
            performance.setMaxDrawdown(maxDrawdown);

            // Calculate accuracy based on prediction vs actual
            BigDecimal accuracy = calculateAccuracy(signal, actualReturn1d);
            performance.setAccuracy(accuracy);

            signalPerformanceRepository.save(performance);

            log.info("Updated performance for signal: {} - 1d: {}, accuracy: {}",
                    signal.getId(), actualReturn1d, accuracy);

        } catch (Exception e) {
            log.error("Failed to update signal performance: {}", e.getMessage());
        }
    }

    /**
     * Calculate signal accuracy
     */
    private BigDecimal calculateAccuracy(TradingSignal signal, BigDecimal actualReturn) {
        if (signal.getActualReturnPercent() == null) {
            return BigDecimal.ZERO;
        }

        // Compare predicted vs actual direction
        boolean predictedUp = "UP".equals(signal.getDirection());
        boolean actualUp = actualReturn.compareTo(BigDecimal.ZERO) > 0;

        if (predictedUp == actualUp) {
            // Correct direction - calculate how close the magnitude was
            BigDecimal predicted = signal.getActualReturnPercent().abs();
            BigDecimal actual = actualReturn.abs();
            BigDecimal diff = predicted.subtract(actual).abs();
            BigDecimal maxPredicted = predicted.max(actual);

            if (maxPredicted.compareTo(BigDecimal.ZERO) == 0) {
                return BigDecimal.ONE;
            }

            BigDecimal accuracy = BigDecimal.ONE.subtract(
                    diff.divide(maxPredicted, 4, RoundingMode.HALF_UP)
            );

            return accuracy.max(BigDecimal.ZERO);
        } else {
            // Wrong direction
            return BigDecimal.ZERO;
        }
    }
}