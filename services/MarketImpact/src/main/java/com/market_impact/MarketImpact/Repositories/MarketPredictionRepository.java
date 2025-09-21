package com.market_impact.MarketImpact.Repositories;

import com.market_impact.MarketImpact.entity.MarketPrediction;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface MarketPredictionRepository extends JpaRepository<MarketPrediction, UUID> {

    List<MarketPrediction> findBySymbol(String symbol);

    List<MarketPrediction> findBySymbolAndDirection(String symbol, String direction);

    List<MarketPrediction> findByArticleId(UUID articleId);

    List<MarketPrediction> findByModelType(String modelType);

    @Query("SELECT mp FROM MarketPrediction mp WHERE mp.symbol = :symbol " +
            "AND mp.predictionTimestamp BETWEEN :startTime AND :endTime " +
            "ORDER BY mp.predictionTimestamp DESC")
    List<MarketPrediction> findBySymbolAndTimestampRange(
            @Param("symbol") String symbol,
            @Param("startTime") LocalDateTime startTime,
            @Param("endTime") LocalDateTime endTime);

    @Query("SELECT mp FROM MarketPrediction mp WHERE mp.predictionTimestamp >= :timestamp " +
            "ORDER BY mp.predictionTimestamp DESC")
    List<MarketPrediction> findRecentPredictions(@Param("timestamp") LocalDateTime timestamp);

    @Query("SELECT mp FROM MarketPrediction mp WHERE mp.confidence >= :minConfidence " +
            "ORDER BY mp.confidence DESC")
    List<MarketPrediction> findHighConfidencePredictions(@Param("minConfidence") double minConfidence);

    @Query("SELECT mp FROM MarketPrediction mp WHERE mp.symbol = :symbol " +
            "ORDER BY mp.predictionTimestamp DESC")
    Page<MarketPrediction> findBySymbolOrderByTimestampDesc(@Param("symbol") String symbol, Pageable pageable);

    @Query("SELECT DISTINCT mp.symbol FROM MarketPrediction mp")
    List<String> findDistinctSymbols();

    Optional<MarketPrediction> findTopBySymbolOrderByPredictionTimestampDesc(String symbol);
}