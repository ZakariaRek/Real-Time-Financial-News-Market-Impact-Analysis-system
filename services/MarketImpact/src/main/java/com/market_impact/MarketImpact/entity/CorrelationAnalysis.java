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
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(name = "correlation_analysis",
        uniqueConstraints = @UniqueConstraint(columnNames = {"symbol", "analysis_date"}))
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CorrelationAnalysis {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(name = "symbol", nullable = false, length = 20)
    @NotBlank
    private String symbol;

    @Column(name = "analysis_date", nullable = false)
    @NotNull
    private LocalDate analysisDate;

    @Column(name = "sentiment_price_correlation", nullable = false, precision = 5, scale = 4)
    @NotNull
    private BigDecimal sentimentPriceCorrelation; // -1.0 to 1.0

    @Column(name = "immediate_correlation", nullable = false, precision = 5, scale = 4)
    @NotNull
    private BigDecimal immediateCorrelation; // -1.0 to 1.0

    @Column(name = "short_term_correlation", nullable = false, precision = 5, scale = 4)
    @NotNull
    private BigDecimal shortTermCorrelation; // -1.0 to 1.0

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    @PrePersist
    public void prePersist() {
        if (analysisDate == null) {
            analysisDate = LocalDate.now();
        }
    }
}