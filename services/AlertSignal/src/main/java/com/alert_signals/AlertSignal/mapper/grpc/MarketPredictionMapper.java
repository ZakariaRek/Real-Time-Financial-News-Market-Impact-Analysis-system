package com.alert_signals.AlertSignal.mapper.grpc;

import com.alert_signals.AlertSignal.dto.grpc.MarketPredictionDto;
import com.alert_signals.AlertSignal.dto.grpc.RiskMetricsDto;
import com.market_impact.grpc.*;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Component
public class MarketPredictionMapper {

    public MarketPredictionDto toDto(MarketPrediction grpcPrediction) {
        if (grpcPrediction == null) {
            return null;
        }

        return MarketPredictionDto.builder()
                .id(UUID.fromString(grpcPrediction.getId().getValue()))
                .articleId(grpcPrediction.hasArticleId() && !grpcPrediction.getArticleId().getValue().isEmpty() ?
                        UUID.fromString(grpcPrediction.getArticleId().getValue()) : null)
                .symbol(grpcPrediction.getSymbol())
                .predictedChangePercent(new BigDecimal(grpcPrediction.getPredictedChangePercent().getValue()))
                .direction(grpcPrediction.getDirection().name())
                .confidence(new BigDecimal(grpcPrediction.getConfidence().getValue()))
                .impactScore(new BigDecimal(grpcPrediction.getImpactScore().getValue()))
                .modelType(grpcPrediction.getModelType())
                .predictionTimestamp(grpcPrediction.hasPredictionTimestamp() ?
                        convertTimestamp(grpcPrediction.getPredictionTimestamp()) : null)
                .riskMetrics(grpcPrediction.getRiskMetricsCount() > 0 ?
                        grpcPrediction.getRiskMetricsList().stream()
                                .filter(this::isValidRiskMetric)  // ADD THIS FILTER
                                .map(this::mapRiskMetrics)
                                .toList() : null)
                .build();
    }

    // ADD THIS HELPER METHOD
    private boolean isValidRiskMetric(RiskMetrics rm) {
        return rm != null
                && rm.hasId()
                && !rm.getId().getValue().isEmpty()
                && rm.hasVar951Day()
                && rm.getVar951Day().getValue() != null
                && !rm.getVar951Day().getValue().isEmpty();
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

    private RiskMetricsDto mapRiskMetrics(RiskMetrics rm) {
        // Helper method to safely parse UUID
        UUID id = (rm.hasId() && !rm.getId().getValue().isEmpty())
                ? UUID.fromString(rm.getId().getValue())
                : null;

        UUID predictionId = (rm.hasPredictionId() && !rm.getPredictionId().getValue().isEmpty())
                ? UUID.fromString(rm.getPredictionId().getValue())
                : null;

        return RiskMetricsDto.builder()
                .id(id)
                .predictionId(predictionId)
                .symbol(rm.getSymbol())
                .var951Day(new BigDecimal(rm.getVar951Day().getValue()))
                .historicalVolatility30d(new BigDecimal(rm.getHistoricalVolatility30D().getValue()))
                .marketCorrelation(new BigDecimal(rm.getMarketCorrelation().getValue()))
                .riskLevel(rm.getRiskLevel().name())
                .build();
    }
}