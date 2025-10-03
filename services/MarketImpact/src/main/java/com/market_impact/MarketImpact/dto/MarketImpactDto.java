// services/MarketImpact/src/main/java/com/market_impact/MarketImpact/dto/MarketImpactDto.java
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
public class MarketImpactDto {
    private String symbol;
    private BigDecimal predictedChangePercent;
    private String direction;
    private BigDecimal confidence;
    private BigDecimal impactScore;
    private String modelType;

    @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
    private LocalDateTime timestamp;
}