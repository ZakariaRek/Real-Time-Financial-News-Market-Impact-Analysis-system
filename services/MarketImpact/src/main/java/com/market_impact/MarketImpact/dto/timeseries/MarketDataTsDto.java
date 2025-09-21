package com.market_impact.MarketImpact.dto.timeseries;

import com.fasterxml.jackson.annotation.JsonFormat;
import jakarta.validation.constraints.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;

public class MarketDataTsDto {

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Response {
        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime timestamp;

        private String symbol;
        private BigDecimal closePrice;
        private Long volume;
        private BigDecimal priceChangePercent;
        private BigDecimal volatility;
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

        @NotNull(message = "Close price is required")
        @DecimalMin(value = "0.0", message = "Close price must be >= 0.0")
        private BigDecimal closePrice;

        @NotNull(message = "Volume is required")
        @Min(value = 0, message = "Volume must be >= 0")
        private Long volume;

        @DecimalMin(value = "-100.0", message = "Price change percent must be >= -100%")
        @DecimalMax(value = "100.0", message = "Price change percent must be <= 100%")
        private BigDecimal priceChangePercent;

        @DecimalMin(value = "0.0", message = "Volatility must be >= 0.0")
        private BigDecimal volatility;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Summary {
        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
        private LocalDateTime timestamp;

        private String symbol;
        private BigDecimal closePrice;
        private BigDecimal priceChangePercent;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Analytics {
        private String symbol;
        private BigDecimal averagePrice;
        private Double averageVolume;
        private BigDecimal volatility;
        private BigDecimal minPrice;
        private BigDecimal maxPrice;
        private LocalDateTime periodStart;
        private LocalDateTime periodEnd;
        private Long dataPoints;
    }
}