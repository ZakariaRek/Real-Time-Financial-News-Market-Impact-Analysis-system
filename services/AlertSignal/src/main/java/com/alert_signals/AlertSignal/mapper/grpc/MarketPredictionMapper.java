package com.alert_signals.AlertSignal.mapper.grpc;

import com.alert_signals.AlertSignal.dto.grpc.MarketPredictionDto;
import com.alert_signals.AlertSignal.dto.grpc.RiskMetricsDto;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Component
public class MarketPredictionMapper {

    public MarketPredictionDto toDto(com.alert_signals.AlertSignal.grpc.generated.MarketPrediction grpcPrediction) {
        if (grpcPrediction == null) {
            return null;
        }

        return MarketPredictionDto.builder()
                .id(UUID.fromString(grpcPrediction.getId().getValue()))
                .articleId(grpcPrediction.getArticleId().getValue().isEmpty() ?
                        null : UUID.fromString(grpcPrediction.getArticleId().getValue()))
                .symbol(grpcPrediction.getSymbol())
                .predictedChangePercent(new BigDecimal(grpcPrediction.getPredictedChangePercent().getValue()))
                .direction(grpcPrediction.getDirection().name())
                .confidence(new BigDecimal(grpcPrediction.getConfidence().getValue()))
                .impactScore(new BigDecimal(grpcPrediction.getImpactScore().getValue()))
                .modelType(grpcPrediction.getModelType())
                .predictionTimestamp(convertTimestamp(grpcPrediction.getPredictionTimestamp()))
                .riskMetrics(mapRiskMetrics(grpcPrediction.getRiskMetricsList()))
                .build();
    }

    private LocalDateTime convertTimestamp(com.google.protobuf.Timestamp timestamp) {
        if (timestamp == null) {
            return null;
        }
        return LocalDateTime.ofEpochSecond(
                timestamp.getSeconds(),
                timestamp.getNanos(),
                ZoneOffset.UTC
        );
    }

    private List<RiskMetricsDto> mapRiskMetrics(
            List<com.alert_signals.AlertSignal.grpc.generated.RiskMetrics> riskMetricsList) {

        if (riskMetricsList == null || riskMetricsList.isEmpty()) {
            return List.of();
        }

        return riskMetricsList.stream()
                .map(rm -> RiskMetricsDto.builder()
                        .id(UUID.fromString(rm.getId().getValue()))
                        .predictionId(UUID.fromString(rm.getPredictionId().getValue()))
                        .symbol(rm.getSymbol())
                        .var951Day(new BigDecimal(rm.getVar951Day().getValue()))
                        .historicalVolatility30d(new BigDecimal(rm.getHistoricalVolatility30D().getValue()))
                        .marketCorrelation(new BigDecimal(rm.getMarketCorrelation().getValue()))
                        .riskLevel(rm.getRiskLevel().name())
                        .build())
                .collect(Collectors.toList());
    }
}