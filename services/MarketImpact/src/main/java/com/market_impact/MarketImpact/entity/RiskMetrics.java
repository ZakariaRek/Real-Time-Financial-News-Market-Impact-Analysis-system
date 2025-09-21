package com.market_impact.MarketImpact.entity;

import jakarta.persistence.*;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(name = "risk_metrics")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class RiskMetrics {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "prediction_id", nullable = false)
    @NotNull
    private MarketPrediction prediction;

    @Column(name = "symbol", nullable = false, length = 20)
    @NotBlank
    private String symbol;

    @Column(name = "var_95_1day", nullable = false, precision = 10, scale = 4)
    @NotNull
    private BigDecimal var951day; // Value at Risk 95% 1-day

    @Column(name = "historical_volatility_30d", nullable = false, precision = 10, scale = 4)
    @NotNull
    private BigDecimal historicalVolatility30d;

    @Column(name = "market_correlation", nullable = false, precision = 5, scale = 4)
    @NotNull
    private BigDecimal marketCorrelation; // -1.0 to 1.0

    @Column(name = "risk_level", nullable = false, length = 20)
    @NotBlank
    private String riskLevel; // LOW, MEDIUM, HIGH, CRITICAL

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private LocalDateTime updatedAt;
}