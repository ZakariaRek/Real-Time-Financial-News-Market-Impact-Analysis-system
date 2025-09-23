package com.market_impact.MarketImpact.Mappers;

import com.market_impact.MarketImpact.dto.MarketPredictionDto;
import com.market_impact.MarketImpact.entity.MarketPrediction;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;

@Component
@RequiredArgsConstructor
public class MarketPredictionMapper {

    private final RiskMetricsMapper riskMetricsMapper;

    public MarketPrediction toEntity(MarketPredictionDto.Request request) {
        if (request == null) {
            return null;
        }

        return MarketPrediction.builder()
                .articleId(request.getArticleId())
                .symbol(request.getSymbol())
                .predictedChangePercent(request.getPredictedChangePercent())
                .direction(request.getDirection())
                .confidence(request.getConfidence())
                .impactScore(request.getImpactScore())
                .modelType(request.getModelType())
                .predictionTimestamp(request.getPredictionTimestamp() != null ?
                        request.getPredictionTimestamp() : LocalDateTime.now())
                .build();
    }

    public MarketPredictionDto.Response toResponse(MarketPrediction entity) {
        if (entity == null) {
            return null;
        }

        return MarketPredictionDto.Response.builder()
                .id(entity.getId())
                .articleId(entity.getArticleId())
                .symbol(entity.getSymbol())
                .predictedChangePercent(entity.getPredictedChangePercent())
                .direction(entity.getDirection())
                .confidence(entity.getConfidence())
                .impactScore(entity.getImpactScore())
                .modelType(entity.getModelType())
                .predictionTimestamp(entity.getPredictionTimestamp())
                .createdAt(entity.getCreatedAt())
                .updatedAt(entity.getUpdatedAt())
                .riskMetrics(entity.getRiskMetrics() != null ?
                        entity.getRiskMetrics().stream()
                                .map(riskMetricsMapper::toResponse)
                                .toList() : null)
                .build();
    }

    public MarketPredictionDto.Summary toSummary(MarketPrediction entity) {
        if (entity == null) {
            return null;
        }

        return MarketPredictionDto.Summary.builder()
                .id(entity.getId())
                .symbol(entity.getSymbol())
                .predictedChangePercent(entity.getPredictedChangePercent())
                .direction(entity.getDirection())
                .confidence(entity.getConfidence())
                .modelType(entity.getModelType())
                .predictionTimestamp(entity.getPredictionTimestamp())
                .build();
    }

    public void updateEntityFromRequest(MarketPredictionDto.Request request, MarketPrediction entity) {
        if (request == null || entity == null) {
            return;
        }

        if (request.getArticleId() != null) {
            entity.setArticleId(request.getArticleId());
        }
        if (request.getSymbol() != null) {
            entity.setSymbol(request.getSymbol());
        }
        if (request.getPredictedChangePercent() != null) {
            entity.setPredictedChangePercent(request.getPredictedChangePercent());
        }
        if (request.getDirection() != null) {
            entity.setDirection(request.getDirection());
        }
        if (request.getConfidence() != null) {
            entity.setConfidence(request.getConfidence());
        }
        if (request.getImpactScore() != null) {
            entity.setImpactScore(request.getImpactScore());
        }
        if (request.getModelType() != null) {
            entity.setModelType(request.getModelType());
        }
        if (request.getPredictionTimestamp() != null) {
            entity.setPredictionTimestamp(request.getPredictionTimestamp());
        }
    }
}