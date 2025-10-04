package com.alert_signals.AlertSignal.client;

import com.alert_signals.AlertSignal.dto.grpc.MarketPredictionDto;
import com.alert_signals.AlertSignal.mapper.grpc.MarketPredictionMapper;
import io.grpc.StatusRuntimeException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.UUID;

@Component
@RequiredArgsConstructor
@Slf4j
public class MarketImpactClient {


    private com.alert_signals.AlertSignal.grpc.generated.MarketPredictionServiceGrpc.MarketPredictionServiceBlockingStub marketPredictionStub;

    private final MarketPredictionMapper mapper;

    public MarketPredictionDto getPrediction(UUID predictionId) {
        try {
            com.alert_signals.AlertSignal.grpc.generated.GetMarketPredictionRequest request =
                    com.alert_signals.AlertSignal.grpc.generated.GetMarketPredictionRequest.newBuilder()
                            .setId(com.alert_signals.AlertSignal.grpc.generated.UUID.newBuilder()
                                    .setValue(predictionId.toString())
                                    .build())
                            .build();

            com.alert_signals.AlertSignal.grpc.generated.GetMarketPredictionResponse response =
                    marketPredictionStub.getPrediction(request);

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
            com.alert_signals.AlertSignal.grpc.generated.GetLatestPredictionRequest request =
                    com.alert_signals.AlertSignal.grpc.generated.GetLatestPredictionRequest.newBuilder()
                            .setSymbol(symbol)
                            .build();

            com.alert_signals.AlertSignal.grpc.generated.GetLatestPredictionResponse response =
                    marketPredictionStub.getLatestPrediction(request);

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
