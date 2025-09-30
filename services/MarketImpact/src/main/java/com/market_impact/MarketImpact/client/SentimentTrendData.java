package com.market_impact.MarketImpact.client;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class SentimentTrendData {
    private LocalDateTime timestamp;
    private String symbol;
    private double avgSentiment;
    private long articleCount;
    private double volatility;
}