package com.alert_signals.AlertSignal.entity;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.GenericGenerator;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(name = "trading_signals")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class TradingSignal {

    @Id
    @GeneratedValue(generator = "UUID")
    @GenericGenerator(name = "UUID", strategy = "org.hibernate.id.UUIDGenerator")
    @Column(name = "id", updatable = false, nullable = false)
    private UUID id;

    @Column(name = "prediction_id")
    private UUID predictionId;

    @Column(name = "symbol", nullable = false)
    private String symbol;

    @Column(name = "signal_type", nullable = false)
    private String signalType;

    @Column(name = "direction", nullable = false)
    private String direction;

    @Column(name = "strength", precision = 10, scale = 4)
    private BigDecimal strength;

    @Column(name = "confidence", precision = 10, scale = 4)
    private BigDecimal confidence;

    @Column(name = "status", nullable = false)
    private String status;

    @CreationTimestamp
    @Column(name = "generated_at", nullable = false)
    private LocalDateTime generatedAt;

    @Column(name = "actual_return_percent", precision = 10, scale = 4)
    private BigDecimal actualReturnPercent;
}
