package com.alert_signals.AlertSignal.grpc;

import com.alert_signals.AlertSignal.entity.TradingSignal;
import com.alert_signals.AlertSignal.service.SignalEvaluationService;
import com.alert_signals.AlertSignal.service.SignalNotificationService;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.UUID;

/**
 * gRPC service to receive prediction impact requests from MarketImpact service
 * This will be properly implemented once proto files are generated
 * For now, it's a placeholder service
 */
@Service
@RequiredArgsConstructor
@Slf4j
public class SignalProcessingGrpcService {

    private final SignalEvaluationService signalEvaluationService;
    private final SignalNotificationService signalNotificationService;

    /**
     * Process prediction impact and generate trading signal if criteria met
     * Called by MarketImpact service (MIS) as shown in diagram
     *
     * This method will be properly implemented once proto files are compiled
     */
    public void processPredictionImpact(String predictionId) {
        try {
            UUID predId = UUID.fromString(predictionId);
            log.info("Received prediction impact request: {}", predId);

            // Process prediction and create signal if high confidence
            TradingSignal signal = signalEvaluationService.processPrediction(predId);

            if (signal != null) {
                log.info("High-confidence signal created: {}", signal.getId());

                // Process notifications
                signalNotificationService.processNewSignal(signal);
            } else {
                log.info("Signal discarded - criteria not met for prediction: {}", predId);
            }

        } catch (Exception e) {
            log.error("Error processing prediction impact: {}", e.getMessage(), e);
            throw new RuntimeException("Error processing prediction: " + e.getMessage());
        }
    }

    /**
     * Process prediction and return result
     */
    public SignalResult processPrediction(String predictionId, String symbol,
                                          double confidence, double impactScore) {
        try {
            UUID predId = UUID.fromString(predictionId);
            TradingSignal signal = signalEvaluationService.processPrediction(predId);

            if (signal != null) {
                signalNotificationService.processNewSignal(signal);

                return SignalResult.builder()
                        .success(true)
                        .signalId(signal.getId().toString())
                        .message("High-confidence signal created")
                        .confidence(signal.getConfidence().doubleValue())
                        .strength(signal.getStrength().doubleValue())
                        .build();
            } else {
                return SignalResult.builder()
                        .success(false)
                        .message("Signal discarded - criteria not met")
                        .build();
            }
        } catch (Exception e) {
            log.error("Error processing prediction: {}", e.getMessage(), e);
            return SignalResult.builder()
                    .success(false)
                    .message("Error: " + e.getMessage())
                    .build();
        }
    }

    /**
     * Simple result class for signal processing
     */
    @lombok.Data
    @lombok.Builder
    public static class SignalResult {
        private boolean success;
        private String signalId;
        private String message;
        private double confidence;
        private double strength;
    }
}