// services/AlertSignal/src/main/java/com/alert_signals/AlertSignal/grpc/SignalProcessingGrpcServiceImpl.java
package com.alert_signals.AlertSignal.grpc;

import com.alert_signals.AlertSignal.entity.TradingSignal;
import com.alert_signals.AlertSignal.grpc.generated.*;
import com.alert_signals.AlertSignal.service.SignalEvaluationService;
import com.alert_signals.AlertSignal.service.SignalNotificationService;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.grpc.server.service.GrpcService;

import java.util.UUID;

@GrpcService
@RequiredArgsConstructor
@Slf4j
public class SignalProcessingGrpcServiceImpl extends SignalProcessingServiceGrpc.SignalProcessingServiceImplBase {

    private final SignalEvaluationService signalEvaluationService;
    private final SignalNotificationService signalNotificationService;

    @Override
    public void predictImpact(PredictImpactRequest request,
                              StreamObserver<PredictImpactResponse> responseObserver) {
        try {
            UUID predictionId = UUID.fromString(request.getPredictionId());
            log.info("Received prediction impact request: {}", predictionId);

            // Process prediction and create signal if criteria met
            TradingSignal signal = signalEvaluationService.processPrediction(predictionId);

            PredictImpactResponse.Builder responseBuilder = PredictImpactResponse.newBuilder();

            if (signal != null) {
                log.info("High-confidence signal created: {}", signal.getId());

                // Process notifications (this broadcasts via WebSocket!)
                signalNotificationService.processNewSignal(signal);

                responseBuilder
                        .setSuccess(true)
                        .setSignalId(signal.getId().toString())
                        .setMessage("High-confidence signal created")
                        .setConfidence(signal.getConfidence().doubleValue())
                        .setStrength(signal.getStrength().doubleValue());
            } else {
                log.info("Signal discarded - criteria not met for prediction: {}", predictionId);
                responseBuilder
                        .setSuccess(false)
                        .setMessage("Signal discarded - criteria not met");
            }

            responseObserver.onNext(responseBuilder.build());
            responseObserver.onCompleted();

        } catch (Exception e) {
            log.error("Error processing prediction impact: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }
}