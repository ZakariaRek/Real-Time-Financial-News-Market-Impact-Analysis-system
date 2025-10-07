// services/MarketImpact/src/main/java/com/market_impact/MarketImpact/client/AlertSignalClient.java
package com.market_impact.MarketImpact.client;

import com.alert_signals.AlertSignal.grpc.generated.SignalProcessingServiceGrpc;
import com.alert_signals.AlertSignal.grpc.generated.PredictImpactRequest;
import com.alert_signals.AlertSignal.grpc.generated.PredictImpactResponse;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

@Component
@Slf4j
public class AlertSignalClient {

    @Value("${external.services.alert-signal.host:localhost}")
    private String host;

    @Value("${external.services.alert-signal.port:9095}")
    private int port;

    private ManagedChannel channel;
    private SignalProcessingServiceGrpc.SignalProcessingServiceBlockingStub stub;

    @PostConstruct
    public void init() {
        log.info("🔌 Initializing AlertSignal gRPC client: {}:{}", host, port);

        this.channel = ManagedChannelBuilder
                .forAddress(host, port)
                .usePlaintext()
                .build();

        this.stub = SignalProcessingServiceGrpc.newBlockingStub(channel);
        log.info("✅ AlertSignal gRPC client initialized successfully");
    }

    @PreDestroy
    public void shutdown() {
        if (channel != null) {
            channel.shutdown();
            log.info("🔌 AlertSignal gRPC client shutdown");
        }
    }

    public void notifyPrediction(String predictionId, String symbol,
                                 double confidence, double impactScore) {
        try {
            log.info("📤 Sending prediction to AlertSignal: predictionId={}, symbol={}, confidence={}, impact={}",
                    predictionId, symbol, confidence, impactScore);

            PredictImpactRequest request = PredictImpactRequest.newBuilder()
                    .setPredictionId(predictionId)
                    .setSymbol(symbol)
                    .setConfidence(confidence)
                    .setImpactScore(impactScore)
                    .build();

            PredictImpactResponse response = stub.predictImpact(request);

            if (response.getSuccess()) {
                log.info("✅ AlertSignal processed prediction successfully: signalId={}, message={}",
                        response.getSignalId(), response.getMessage());
            } else {
                log.warn("⚠️ AlertSignal did not create signal: {}", response.getMessage());
            }
        } catch (Exception e) {
            log.error("❌ Failed to notify AlertSignal: {} - {}", e.getClass().getName(), e.getMessage());
            e.printStackTrace();
        }
    }
}