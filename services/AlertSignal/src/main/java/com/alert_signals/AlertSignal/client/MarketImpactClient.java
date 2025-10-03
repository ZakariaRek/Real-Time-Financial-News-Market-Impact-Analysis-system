package com.alert_signals.AlertSignal.client;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.UUID;

@Component
@Slf4j
public class MarketImpactClient {

    @Value("${external.services.market-impact.host:localhost}")
    private String marketImpactHost;

    @Value("${external.services.market-impact.port:9091}")
    private int marketImpactPort;

    private ManagedChannel channel;
    private com.alert_signals.AlertSignal.grpc.generated.MarketPredictionServiceGrpc.MarketPredictionServiceBlockingStub blockingStub;

    @PostConstruct
    public void init() {
        try {
            this.channel = ManagedChannelBuilder
                    .forAddress(marketImpactHost, marketImpactPort)
                    .usePlaintext()
                    .build();

            this.blockingStub = com.alert_signals.AlertSignal.grpc.generated.MarketPredictionServiceGrpc.newBlockingStub(channel);
            log.info("MarketImpact client initialized: {}:{}", marketImpactHost, marketImpactPort);
        } catch (Exception e) {
            log.error("Failed to initialize MarketImpact client: {}", e.getMessage(), e);
        }
    }

    @PreDestroy
    public void shutdown() {
        if (channel != null && !channel.isShutdown()) {
            channel.shutdown();
            log.info("MarketImpact client shutdown");
        }
    }

    public com.alert_signals.AlertSignal.grpc.generated.MarketPrediction getPrediction(UUID predictionId) {
        try {
            com.alert_signals.AlertSignal.grpc.generated.GetMarketPredictionRequest request = com.alert_signals.AlertSignal.grpc.generated.GetMarketPredictionRequest.newBuilder()
                    .setId(com.alert_signals.AlertSignal.grpc.generated.UUID.newBuilder()
                            .setValue(predictionId.toString())
                            .build())
                    .build();

            com.alert_signals.AlertSignal.grpc.generated.GetMarketPredictionResponse response = blockingStub.getPrediction(request);
            return response.getPrediction();
        } catch (StatusRuntimeException e) {
            log.error("gRPC error getting prediction {}: {}", predictionId, e.getMessage());
            throw new RuntimeException("Failed to get prediction: " + e.getMessage());
        }
    }

    public com.alert_signals.AlertSignal.grpc.generated.MarketPrediction getLatestPrediction(String symbol) {
        try {
            com.alert_signals.AlertSignal.grpc.generated.GetLatestPredictionRequest request = com.alert_signals.AlertSignal.grpc.generated.GetLatestPredictionRequest.newBuilder()
                    .setSymbol(symbol)
                    .build();

            com.alert_signals.AlertSignal.grpc.generated.GetLatestPredictionResponse response = blockingStub.getLatestPrediction(request);
            return response.getPrediction();
        } catch (StatusRuntimeException e) {
            log.error("gRPC error getting latest prediction for {}: {}", symbol, e.getMessage());
            throw new RuntimeException("Failed to get latest prediction: " + e.getMessage());
        }
    }
}