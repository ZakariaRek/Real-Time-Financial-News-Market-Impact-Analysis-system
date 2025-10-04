package com.alert_signals.AlertSignal.dto.grpc;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class RiskMetricsDto {
    private UUID id;
    private UUID predictionId;
    private String symbol;
    private BigDecimal var951Day;
    private BigDecimal historicalVolatility30d;
    private BigDecimal marketCorrelation;
    private String riskLevel;
}