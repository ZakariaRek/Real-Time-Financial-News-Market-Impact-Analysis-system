package com.market_impact.MarketImpact.dto;

import com.fasterxml.jackson.annotation.JsonFormat;
import jakarta.validation.constraints.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.UUID;

public class CorrelationAnalysisDto {

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Response {
        private UUID id;
        private String symbol;

        @JsonFormat(pattern = "yyyy-MM-dd")
        private LocalDate analysisDate;

        private BigDecimal sentimentPriceCorrelation;
        private BigDecimal immediateCorrelation;
        private BigDecimal shortTermCorrelation;

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
        @NotBlank(message = "Symbol is required")
        @Size(max = 20, message = "Symbol must not exceed 20 characters")
        private String symbol;

        @JsonFormat(pattern = "yyyy-MM-dd")
        private LocalDate analysisDate;

        @NotNull(message = "Sentiment price correlation is required")
        @DecimalMin(value = "-1.0", message = "Sentiment price correlation must be >= -1.0")
        @DecimalMax(value = "1.0", message = "Sentiment price correlation must be <= 1.0")
        private BigDecimal sentimentPriceCorrelation;

        @NotNull(message = "Immediate correlation is required")
        @DecimalMin(value = "-1.0", message = "Immediate correlation must be >= -1.0")
        @DecimalMax(value = "1.0", message = "Immediate correlation must be <= 1.0")
        private BigDecimal immediateCorrelation;

        @NotNull(message = "Short term correlation is required")
        @DecimalMin(value = "-1.0", message = "Short term correlation must be >= -1.0")
        @DecimalMax(value = "1.0", message = "Short term correlation must be <= 1.0")
        private BigDecimal shortTermCorrelation;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Summary {
        private UUID id;
        private String symbol;

        @JsonFormat(pattern = "yyyy-MM-dd")
        private LocalDate analysisDate;

        private BigDecimal sentimentPriceCorrelation;
        private BigDecimal immediateCorrelation;
        private BigDecimal shortTermCorrelation;
    }
}