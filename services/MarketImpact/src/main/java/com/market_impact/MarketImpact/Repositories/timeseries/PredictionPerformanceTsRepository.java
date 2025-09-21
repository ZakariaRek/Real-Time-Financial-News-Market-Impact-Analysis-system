package com.market_impact.MarketImpact.Repositories.timeseries;

import com.market_impact.MarketImpact.entity.timeseries.PredictionPerformanceTs;
import com.market_impact.MarketImpact.entity.timeseries.PredictionPerformanceTsId;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

@Repository
public interface PredictionPerformanceTsRepository extends JpaRepository<PredictionPerformanceTs, PredictionPerformanceTsId> {

    List<PredictionPerformanceTs> findBySymbol(String symbol);

    List<PredictionPerformanceTs> findByModelVersion(String modelVersion);

    List<PredictionPerformanceTs> findBySymbolAndModelVersion(String symbol, String modelVersion);

    @Query("SELECT pp FROM PredictionPerformanceTs pp WHERE pp.symbol = :symbol " +
            "AND pp.modelVersion = :modelVersion " +
            "AND pp.timestamp BETWEEN :startTime AND :endTime " +
            "ORDER BY pp.timestamp DESC")
    List<PredictionPerformanceTs> findBySymbolModelAndTimeRange(
            @Param("symbol") String symbol,
            @Param("modelVersion") String modelVersion,
            @Param("startTime") LocalDateTime startTime,
            @Param("endTime") LocalDateTime endTime);

    @Query("SELECT pp FROM PredictionPerformanceTs pp WHERE pp.timestamp >= :timestamp " +
            "ORDER BY pp.timestamp DESC")
    List<PredictionPerformanceTs> findRecentPerformance(@Param("timestamp") LocalDateTime timestamp);

    @Query("SELECT pp FROM PredictionPerformanceTs pp WHERE pp.accuracyRate >= :threshold " +
            "ORDER BY pp.accuracyRate DESC")
    List<PredictionPerformanceTs> findHighAccuracyPerformance(@Param("threshold") double threshold);

    @Query("SELECT pp FROM PredictionPerformanceTs pp WHERE pp.symbol = :symbol " +
            "AND pp.modelVersion = :modelVersion " +
            "ORDER BY pp.timestamp DESC LIMIT 1")
    Optional<PredictionPerformanceTs> findLatestBySymbolAndModel(
            @Param("symbol") String symbol,
            @Param("modelVersion") String modelVersion);

    @Query("SELECT AVG(pp.accuracyRate) FROM PredictionPerformanceTs pp " +
            "WHERE pp.symbol = :symbol AND pp.modelVersion = :modelVersion " +
            "AND pp.timestamp BETWEEN :startTime AND :endTime")
    BigDecimal getAverageAccuracyRate(
            @Param("symbol") String symbol,
            @Param("modelVersion") String modelVersion,
            @Param("startTime") LocalDateTime startTime,
            @Param("endTime") LocalDateTime endTime);

    @Query("SELECT AVG(pp.sharpeRatio) FROM PredictionPerformanceTs pp " +
            "WHERE pp.symbol = :symbol AND pp.modelVersion = :modelVersion " +
            "AND pp.timestamp BETWEEN :startTime AND :endTime")
    BigDecimal getAverageSharpeRatio(
            @Param("symbol") String symbol,
            @Param("modelVersion") String modelVersion,
            @Param("startTime") LocalDateTime startTime,
            @Param("endTime") LocalDateTime endTime);

    @Query("SELECT AVG(pp.winRate) FROM PredictionPerformanceTs pp " +
            "WHERE pp.symbol = :symbol AND pp.modelVersion = :modelVersion " +
            "AND pp.timestamp BETWEEN :startTime AND :endTime")
    BigDecimal getAverageWinRate(
            @Param("symbol") String symbol,
            @Param("modelVersion") String modelVersion,
            @Param("startTime") LocalDateTime startTime,
            @Param("endTime") LocalDateTime endTime);

    @Query("SELECT DISTINCT pp.symbol FROM PredictionPerformanceTs pp")
    List<String> findDistinctSymbols();

    @Query("SELECT DISTINCT pp.modelVersion FROM PredictionPerformanceTs pp")
    List<String> findDistinctModelVersions();
}