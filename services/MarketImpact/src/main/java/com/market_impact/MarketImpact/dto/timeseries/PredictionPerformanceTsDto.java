package com.market_impact.MarketImpact.dto.timeseries;

import com.fasterxml.jackson.annotation.JsonFormat;
import jakarta.validation.constraints.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;

public class PredictionPerformanceTsDto {

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Response {
        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime timestamp;

        private String symbol;
        private String modelVersion;
        private BigDecimal accuracyRate;
        private BigDecimal sharpeRatio;
        private BigDecimal winRate;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Request {
        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime timestamp;

        @NotBlank(message = "Symbol is required")
        @Size(max = 20, message = "Symbol must not exceed 20 characters")
        private String symbol;

        @NotBlank(message = "Model version is required")
        @Size(max = 50, message = "Model version must not exceed 50 characters")
        private String modelVersion;

        @NotNull(message = "Accuracy rate is required")
        @DecimalMin(value = "0.0", message = "Accuracy rate must be >= 0.0")
        @DecimalMax(value = "1.0", message = "Accuracy rate must be <= 1.0")
        private BigDecimal accuracyRate;

        private BigDecimal sharpeRatio;

        @DecimalMin(value = "0.0", message = "Win rate must be >= 0.0")
        @DecimalMax(value = "1.0", message = "Win rate must be <= 1.0")
        private BigDecimal winRate;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Summary {
        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime timestamp;

        private String symbol;
        private String modelVersion;
        private BigDecimal accuracyRate;
        private BigDecimal winRate;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Analytics {
        private String symbol;
        private String modelVersion;
        private BigDecimal averageAccuracyRate;
        private BigDecimal averageSharpeRatio;
        private BigDecimal averageWinRate;
        private BigDecimal maxAccuracyRate;
        private BigDecimal minAccuracyRate;
        private LocalDateTime periodStart;
        private LocalDateTime periodEnd;
        private Long dataPoints;
    }
}