package com.market_impact.MarketImpact.entity.timeseries;

import jakarta.persistence.*;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Entity
@Table(name = "market_data_ts")
@IdClass(MarketDataTsId.class)
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class MarketDataTs {

    @Id
    @Column(name = "timestamp", nullable = false)
    @NotNull
    private LocalDateTime timestamp;

    @Id
    @Column(name = "symbol", nullable = false, length = 20)
    @NotBlank
    private String symbol;

    @Column(name = "close_price", nullable = false, precision = 12, scale = 4)
    @NotNull
    private BigDecimal closePrice;

    @Column(name = "volume", nullable = false)
    @NotNull
    private Long volume;

    @Column(name = "price_change_percent", precision = 8, scale = 4)
    private BigDecimal priceChangePercent;

    @Column(name = "volatility", precision = 8, scale = 4)
    private BigDecimal volatility;

    @PrePersist
    public void prePersist() {
        if (timestamp == null) {
            timestamp = LocalDateTime.now();
        }
    }
}