package com.alert_signals.AlertSignal.entity;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.GenericGenerator;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.UUID;

@Entity
@Table(name = "signal_performance",
        uniqueConstraints = @UniqueConstraint(columnNames = {"signal_id", "performance_date"}))
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class SignalPerformance {

    @Id
    @GeneratedValue(generator = "UUID")
    @GenericGenerator(name = "UUID", strategy = "org.hibernate.id.UUIDGenerator")
    @Column(name = "id", updatable = false, nullable = false)
    private UUID id;

    @Column(name = "signal_id", nullable = false)
    private UUID signalId;

    @Column(name = "performance_date", nullable = false)
    private LocalDate performanceDate;

    @Column(name = "return_1d", precision = 10, scale = 4)
    private BigDecimal return1d;

    @Column(name = "return_1w", precision = 10, scale = 4)
    private BigDecimal return1w;

    @Column(name = "max_drawdown", precision = 10, scale = 4)
    private BigDecimal maxDrawdown;

    @Column(name = "accuracy", precision = 10, scale = 4)
    private BigDecimal accuracy;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "signal_id", insertable = false, updatable = false)
    private TradingSignal tradingSignal;
}
