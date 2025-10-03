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
public class SentimentData {
    private String articleId;
    private float compoundScore;
    private float confidence;
    private String primarySymbol;
    private LocalDateTime analysisTimestamp;
}