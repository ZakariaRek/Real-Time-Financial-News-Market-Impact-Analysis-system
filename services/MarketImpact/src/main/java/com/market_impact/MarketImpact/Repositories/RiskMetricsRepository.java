package com.market_impact.MarketImpact.Repositories;

import com.market_impact.MarketImpact.entity.RiskMetrics;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface RiskMetricsRepository extends JpaRepository<RiskMetrics, UUID> {

    List<RiskMetrics> findBySymbol(String symbol);

    List<RiskMetrics> findByRiskLevel(String riskLevel);

    List<RiskMetrics> findByPredictionId(UUID predictionId);

    @Query("SELECT rm FROM RiskMetrics rm WHERE rm.symbol = :symbol " +
            "AND rm.riskLevel = :riskLevel ORDER BY rm.createdAt DESC")
    List<RiskMetrics> findBySymbolAndRiskLevel(@Param("symbol") String symbol,
                                               @Param("riskLevel") String riskLevel);

    @Query("SELECT rm FROM RiskMetrics rm WHERE rm.var951day >= :threshold " +
            "ORDER BY rm.var951day DESC")
    List<RiskMetrics> findHighVaRMetrics(@Param("threshold") double threshold);

    @Query("SELECT rm FROM RiskMetrics rm WHERE rm.historicalVolatility30d >= :threshold " +
            "ORDER BY rm.historicalVolatility30d DESC")
    List<RiskMetrics> findHighVolatilityMetrics(@Param("threshold") double threshold);

    @Query("SELECT DISTINCT rm.symbol FROM RiskMetrics rm WHERE rm.riskLevel = :riskLevel")
    List<String> findSymbolsByRiskLevel(@Param("riskLevel") String riskLevel);

    @Query("SELECT AVG(rm.marketCorrelation) FROM RiskMetrics rm WHERE rm.symbol = :symbol")
    Double getAverageMarketCorrelationBySymbol(@Param("symbol") String symbol);
}