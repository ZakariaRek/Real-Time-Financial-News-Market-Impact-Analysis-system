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
@Table(name = "prediction_performance_ts")
@IdClass(PredictionPerformanceTsId.class)
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class PredictionPerformanceTs {

    @Id
    @Column(name = "timestamp", nullable = false)
    @NotNull
    private LocalDateTime timestamp;

    @Id
    @Column(name = "symbol", nullable = false, length = 20)
    @NotBlank
    private String symbol;

    @Id
    @Column(name = "model_version", nullable = false, length = 50)
    @NotBlank
    private String modelVersion;

    @Column(name = "accuracy_rate", nullable = false, precision = 5, scale = 4)
    @NotNull
    private BigDecimal accuracyRate; // 0.0 to 1.0

    @Column(name = "sharpe_ratio", precision = 8, scale = 4)
    private BigDecimal sharpeRatio;

    @Column(name = "win_rate", precision = 5, scale = 4)
    private BigDecimal winRate; // 0.0 to 1.0

    @PrePersist
    public void prePersist() {
        if (timestamp == null) {
            timestamp = LocalDateTime.now();
        }
    }
}