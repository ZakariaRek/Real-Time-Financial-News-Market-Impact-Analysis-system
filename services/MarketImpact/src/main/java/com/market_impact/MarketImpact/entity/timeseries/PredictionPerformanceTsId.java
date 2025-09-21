package com.market_impact.MarketImpact.entity.timeseries;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.io.Serializable;
import java.time.LocalDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class PredictionPerformanceTsId implements Serializable {
    private LocalDateTime timestamp;
    private String symbol;
    private String modelVersion;
}