// services/MarketImpact/src/main/java/com/market_impact/MarketImpact/dto/MarketSentimentSummary.java
package com.market_impact.MarketImpact.dto;

import com.fasterxml.jackson.annotation.JsonFormat;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class MarketSentimentSummary {

    @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
    private LocalDateTime timestamp;

    private int totalStocks;
    private long bullishCount;
    private long bearishCount;
    private long neutralCount;

    private double bullishPercentage;
    private double bearishPercentage;

    private BigDecimal averageImpactScore;
    private BigDecimal averageConfidence;

    private String marketSentiment; // BULLISH, BEARISH, NEUTRAL
}