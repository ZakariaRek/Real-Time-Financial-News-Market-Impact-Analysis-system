package com.market_impact.MarketImpact.Repositories;

import com.market_impact.MarketImpact.entity.CorrelationAnalysis;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface CorrelationAnalysisRepository extends JpaRepository<CorrelationAnalysis, UUID> {

    List<CorrelationAnalysis> findBySymbol(String symbol);

    Optional<CorrelationAnalysis> findBySymbolAndAnalysisDate(String symbol, LocalDate analysisDate);

    List<CorrelationAnalysis> findByAnalysisDate(LocalDate analysisDate);

    @Query("SELECT ca FROM CorrelationAnalysis ca WHERE ca.symbol = :symbol " +
            "AND ca.analysisDate BETWEEN :startDate AND :endDate " +
            "ORDER BY ca.analysisDate DESC")
    List<CorrelationAnalysis> findBySymbolAndDateRange(
            @Param("symbol") String symbol,
            @Param("startDate") LocalDate startDate,
            @Param("endDate") LocalDate endDate);

    @Query("SELECT ca FROM CorrelationAnalysis ca WHERE ca.analysisDate >= :date " +
            "ORDER BY ca.analysisDate DESC")
    List<CorrelationAnalysis> findRecentAnalysis(@Param("date") LocalDate date);

    @Query("SELECT ca FROM CorrelationAnalysis ca WHERE " +
            "ABS(ca.sentimentPriceCorrelation) >= :threshold " +
            "ORDER BY ABS(ca.sentimentPriceCorrelation) DESC")
    List<CorrelationAnalysis> findHighSentimentCorrelation(@Param("threshold") double threshold);

    @Query("SELECT ca FROM CorrelationAnalysis ca WHERE " +
            "ABS(ca.immediateCorrelation) >= :threshold " +
            "ORDER BY ABS(ca.immediateCorrelation) DESC")
    List<CorrelationAnalysis> findHighImmediateCorrelation(@Param("threshold") double threshold);

    @Query("SELECT ca FROM CorrelationAnalysis ca WHERE " +
            "ABS(ca.shortTermCorrelation) >= :threshold " +
            "ORDER BY ABS(ca.shortTermCorrelation) DESC")
    List<CorrelationAnalysis> findHighShortTermCorrelation(@Param("threshold") double threshold);

    @Query("SELECT AVG(ca.sentimentPriceCorrelation) FROM CorrelationAnalysis ca " +
            "WHERE ca.symbol = :symbol AND ca.analysisDate >= :date")
    Double getAverageSentimentCorrelation(@Param("symbol") String symbol, @Param("date") LocalDate date);

    Optional<CorrelationAnalysis> findTopBySymbolOrderByAnalysisDateDesc(String symbol);
}