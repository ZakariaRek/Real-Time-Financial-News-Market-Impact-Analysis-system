package com.alert_signals.AlertSignal.dto.grpc;

import com.alert_signals.AlertSignal.dto.grpc.RiskMetricsDto;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class MarketPredictionDto {
    private UUID id;
    private UUID articleId;
    private String symbol;
    private BigDecimal predictedChangePercent;
    private String direction;
    private BigDecimal confidence;
    private BigDecimal impactScore;
    private String modelType;
    private LocalDateTime predictionTimestamp;
    private List<RiskMetricsDto> riskMetrics;
}