package com.market_impact.MarketImpact.dto;

import com.fasterxml.jackson.annotation.JsonFormat;
import jakarta.validation.constraints.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

public class MarketPredictionDto {

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Response {
        private UUID id;
        private UUID articleId;
        private String symbol;
        private BigDecimal predictedChangePercent;
        private String direction;
        private BigDecimal confidence;
        private BigDecimal impactScore;
        private String modelType;

        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime predictionTimestamp;

        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime createdAt;

        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime updatedAt;

        private List<RiskMetricsDto.Response> riskMetrics;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Request {

        private UUID articleId;

        @NotBlank(message = "Symbol is required")
        @Size(max = 20, message = "Symbol must not exceed 20 characters")
        private String symbol;

        @NotNull(message = "Predicted change percent is required")
        @DecimalMin(value = "-100.0", message = "Predicted change percent must be >= -100%")
        @DecimalMax(value = "100.0", message = "Predicted change percent must be <= 100%")
        private BigDecimal predictedChangePercent;

        @NotBlank(message = "Direction is required")
        @Pattern(regexp = "UP|DOWN|NEUTRAL", message = "Direction must be UP, DOWN, or NEUTRAL")
        private String direction;

        @NotNull(message = "Confidence is required")
        @DecimalMin(value = "0.0", message = "Confidence must be >= 0.0")
        @DecimalMax(value = "1.0", message = "Confidence must be <= 1.0")
        private BigDecimal confidence;

        @NotNull(message = "Impact score is required")
        private BigDecimal impactScore;

        @NotBlank(message = "Model type is required")
        @Size(max = 50, message = "Model type must not exceed 50 characters")
        private String modelType;

        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime predictionTimestamp;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Summary {
        private UUID id;
        private String symbol;
        private BigDecimal predictedChangePercent;
        private String direction;
        private BigDecimal confidence;
        private String modelType;

        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime predictionTimestamp;
    }
}