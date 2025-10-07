package com.alert_signals.AlertSignal.client;

import com.alert_signals.AlertSignal.dto.grpc.MarketPredictionDto;
import com.alert_signals.AlertSignal.mapper.grpc.MarketPredictionMapper;
import com.market_impact.grpc.*; // Changed to use MarketImpact's generated classes
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.UUID;

@Component
@RequiredArgsConstructor
@Slf4j
public class MarketImpactClient {

    @Value("${grpc.client.market-impact-service.address:static://localhost:9090}")
    private String address;

    private ManagedChannel channel;
    private MarketPredictionServiceGrpc.MarketPredictionServiceBlockingStub marketPredictionStub;

    private final MarketPredictionMapper mapper;

    @PostConstruct
    public void init() {
        try {
            String[] parts = address.replace("static://", "").split(":");
            String host = parts[0];
            int port = Integer.parseInt(parts[1]);

            this.channel = ManagedChannelBuilder
                    .forAddress(host, port)
                    .usePlaintext()
                    .build();

            this.marketPredictionStub = MarketPredictionServiceGrpc.newBlockingStub(channel);

            log.info("MarketImpact gRPC client initialized: {}:{}", host, port);
        } catch (Exception e) {
            log.error("Failed to initialize MarketImpact gRPC client: {}", e.getMessage(), e);
        }
    }

    @PreDestroy
    public void shutdown() {
        if (channel != null && !channel.isShutdown()) {
            channel.shutdown();
            log.info("MarketImpact gRPC client shutdown");
        }
    }

    public MarketPredictionDto getPrediction(UUID predictionId) {
        try {
            GetMarketPredictionRequest request = GetMarketPredictionRequest.newBuilder()
                    .setId(com.market_impact.grpc.UUID.newBuilder()
                            .setValue(predictionId.toString())
                            .build())
                    .build();

            GetMarketPredictionResponse response = marketPredictionStub.getPrediction(request);

            log.info("Successfully retrieved prediction for ID: {}", predictionId);
            return mapper.toDto(response.getPrediction());

        } catch (StatusRuntimeException e) {
            log.error("gRPC error getting prediction {}: {}", predictionId, e.getStatus());
            throw new RuntimeException("Failed to get prediction: " + e.getMessage(), e);
        } catch (Exception e) {
            log.error("Unexpected error getting prediction {}: {}", predictionId, e.getMessage());
            throw new RuntimeException("Failed to get prediction: " + e.getMessage(), e);
        }
    }

    public MarketPredictionDto getLatestPrediction(String symbol) {
        try {
            GetLatestPredictionRequest request = GetLatestPredictionRequest.newBuilder()
                    .setSymbol(symbol)
                    .build();

            GetLatestPredictionResponse response = marketPredictionStub.getLatestPrediction(request);

            log.info("Successfully retrieved latest prediction for symbol: {}", symbol);
            return mapper.toDto(response.getPrediction());

        } catch (StatusRuntimeException e) {
            log.error("gRPC error getting latest prediction for {}: {}", symbol, e.getStatus());
            throw new RuntimeException("Failed to get latest prediction: " + e.getMessage(), e);
        } catch (Exception e) {
            log.error("Unexpected error getting latest prediction for {}: {}", symbol, e.getMessage());
            throw new RuntimeException("Failed to get latest prediction: " + e.getMessage(), e);
        }
    }
}