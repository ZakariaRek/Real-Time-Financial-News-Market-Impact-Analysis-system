package com.market_impact.MarketImpact.dto;

import com.fasterxml.jackson.annotation.JsonFormat;
import jakarta.validation.constraints.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.UUID;

public class RiskMetricsDto {

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Response {
        private UUID id;
        private UUID predictionId;
        private String symbol;
        private BigDecimal var951day;
        private BigDecimal historicalVolatility30d;
        private BigDecimal marketCorrelation;
        private String riskLevel;

        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime createdAt;

        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime updatedAt;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Request {
        @NotNull(message = "Prediction ID is required")
        private UUID predictionId;

        @NotBlank(message = "Symbol is required")
        @Size(max = 20, message = "Symbol must not exceed 20 characters")
        private String symbol;

        @NotNull(message = "VaR 95% 1-day is required")
        @DecimalMin(value = "0.0", message = "VaR must be >= 0.0")
        private BigDecimal var951day;

        @NotNull(message = "Historical volatility 30d is required")
        @DecimalMin(value = "0.0", message = "Historical volatility must be >= 0.0")
        private BigDecimal historicalVolatility30d;

        @NotNull(message = "Market correlation is required")
        @DecimalMin(value = "-1.0", message = "Market correlation must be >= -1.0")
        @DecimalMax(value = "1.0", message = "Market correlation must be <= 1.0")
        private BigDecimal marketCorrelation;

        @NotBlank(message = "Risk level is required")
        @Pattern(regexp = "LOW|MEDIUM|HIGH|CRITICAL", message = "Risk level must be LOW, MEDIUM, HIGH, or CRITICAL")
        private String riskLevel;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Summary {
        private UUID id;
        private String symbol;
        private BigDecimal var951day;
        private String riskLevel;

        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime createdAt;
    }
}