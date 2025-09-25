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
import java.util.List;
import java.util.UUID;

@Entity
@Table(name = "market_predictions")
@Data  // This generates getters, setters, equals, hashCode, toString
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class MarketPrediction {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(name = "article_id", nullable = false)
    @NotNull
    private UUID articleId;

    @Column(name = "symbol", nullable = false, length = 20)
    @NotBlank
    private String symbol;

    @Column(name = "predicted_change_percent", nullable = false, precision = 10, scale = 4)
    @NotNull
    private BigDecimal predictedChangePercent;

    @Column(name = "direction", nullable = false, length = 10)
    @NotBlank
    private String direction; // UP, DOWN, NEUTRAL

    @Column(name = "confidence", nullable = false, precision = 5, scale = 4)
    @NotNull
    private BigDecimal confidence; // 0.0 to 1.0

    @Column(name = "impact_score", nullable = false, precision = 10, scale = 4)
    @NotNull
    private BigDecimal impactScore;

    @Column(name = "model_type", nullable = false, length = 50)
    @NotBlank
    private String modelType;

    @Column(name = "prediction_timestamp", nullable = false)
    @NotNull
    private LocalDateTime predictionTimestamp;

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    @OneToMany(mappedBy = "prediction", cascade = CascadeType.ALL, fetch = FetchType.LAZY)
    private List<RiskMetrics> riskMetrics;

    public UUID getId() {
        return id;
    }

    public void setId(UUID id) {
        this.id = id;
    }

    public UUID getArticleId() {
        return articleId;
    }

    public void setArticleId(UUID articleId) {
        this.articleId = articleId;
    }

    public String getSymbol() {
        return symbol;
    }

    public void setSymbol(String symbol) {
        this.symbol = symbol;
    }

    public BigDecimal getPredictedChangePercent() {
        return predictedChangePercent;
    }

    public void setPredictedChangePercent(BigDecimal predictedChangePercent) {
        this.predictedChangePercent = predictedChangePercent;
    }

    public String getDirection() {
        return direction;
    }

    public void setDirection(String direction) {
        this.direction = direction;
    }

    public BigDecimal getConfidence() {
        return confidence;
    }

    public void setConfidence(BigDecimal confidence) {
        this.confidence = confidence;
    }

    public BigDecimal getImpactScore() {
        return impactScore;
    }

    public void setImpactScore(BigDecimal impactScore) {
        this.impactScore = impactScore;
    }

    public String getModelType() {
        return modelType;
    }

    public void setModelType(String modelType) {
        this.modelType = modelType;
    }

    public LocalDateTime getPredictionTimestamp() {
        return predictionTimestamp;
    }

    public void setPredictionTimestamp(LocalDateTime predictionTimestamp) {
        this.predictionTimestamp = predictionTimestamp;
    }

    public LocalDateTime getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(LocalDateTime createdAt) {
        this.createdAt = createdAt;
    }

    public LocalDateTime getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(LocalDateTime updatedAt) {
        this.updatedAt = updatedAt;
    }

    public List<RiskMetrics> getRiskMetrics() {
        return riskMetrics;
    }

    public void setRiskMetrics(List<RiskMetrics> riskMetrics) {
        this.riskMetrics = riskMetrics;
    }

    @PrePersist
    public void prePersist() {
        if (predictionTimestamp == null) {
            predictionTimestamp = LocalDateTime.now();
        }
    }
}