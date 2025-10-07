// services/MarketImpact/src/main/java/com/market_impact/MarketImpact/controller/PredictionRequest.java
package com.market_impact.MarketImpact.controller;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.UUID;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class PredictionRequest {
    private UUID articleId;
    private String symbol;
}