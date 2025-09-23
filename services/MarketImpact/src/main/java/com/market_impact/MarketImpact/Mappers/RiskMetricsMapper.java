package com.market_impact.MarketImpact.Mappers;

import com.market_impact.MarketImpact.dto.RiskMetricsDto;
import com.market_impact.MarketImpact.entity.RiskMetrics;
import org.springframework.stereotype.Component;

@Component
public class RiskMetricsMapper {

    public RiskMetricsDto.Response toResponse(RiskMetrics entity) {
        if (entity == null) {
            return null;
        }

        return RiskMetricsDto.Response.builder()
                .id(entity.getId())
                .predictionId(entity.getPrediction() != null ? entity.getPrediction().getId() : null)
                .symbol(entity.getSymbol())
                .var951day(entity.getVar951day())
                .historicalVolatility30d(entity.getHistoricalVolatility30d())
                .marketCorrelation(entity.getMarketCorrelation())
                .riskLevel(entity.getRiskLevel())
                .createdAt(entity.getCreatedAt())
                .updatedAt(entity.getUpdatedAt())
                .build();
    }

    public RiskMetricsDto.Summary toSummary(RiskMetrics entity) {
        if (entity == null) {
            return null;
        }

        return RiskMetricsDto.Summary.builder()
                .id(entity.getId())
                .symbol(entity.getSymbol())
                .var951day(entity.getVar951day())
                .riskLevel(entity.getRiskLevel())
                .createdAt(entity.getCreatedAt())
                .build();
    }
}