package com.market_impact.MarketImpact.Repositories.timeseries;

import com.market_impact.MarketImpact.entity.timeseries.MarketDataTs;
import com.market_impact.MarketImpact.entity.timeseries.MarketDataTsId;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

@Repository
public interface MarketDataTsRepository extends JpaRepository<MarketDataTs, MarketDataTsId> {

    List<MarketDataTs> findBySymbol(String symbol);

    @Query("SELECT md FROM MarketDataTs md WHERE md.symbol = :symbol " +
            "AND md.timestamp BETWEEN :startTime AND :endTime " +
            "ORDER BY md.timestamp DESC")
    List<MarketDataTs> findBySymbolAndTimeRange(
            @Param("symbol") String symbol,
            @Param("startTime") LocalDateTime startTime,
            @Param("endTime") LocalDateTime endTime);

    @Query("SELECT md FROM MarketDataTs md WHERE md.timestamp >= :timestamp " +
            "ORDER BY md.timestamp DESC")
    List<MarketDataTs> findRecentData(@Param("timestamp") LocalDateTime timestamp);

    @Query("SELECT md FROM MarketDataTs md WHERE md.symbol = :symbol " +
            "ORDER BY md.timestamp DESC LIMIT 1")
    Optional<MarketDataTs> findLatestBySymbol(@Param("symbol") String symbol);

    @Query("SELECT md FROM MarketDataTs md WHERE md.symbol = :symbol " +
            "ORDER BY md.timestamp DESC LIMIT :limit")
    List<MarketDataTs> findLatestBySymbolLimit(@Param("symbol") String symbol, @Param("limit") int limit);

    @Query("SELECT AVG(md.closePrice) FROM MarketDataTs md WHERE md.symbol = :symbol " +
            "AND md.timestamp BETWEEN :startTime AND :endTime")
    BigDecimal getAveragePriceInTimeRange(
            @Param("symbol") String symbol,
            @Param("startTime") LocalDateTime startTime,
            @Param("endTime") LocalDateTime endTime);

    @Query("SELECT AVG(md.volume) FROM MarketDataTs md WHERE md.symbol = :symbol " +
            "AND md.timestamp BETWEEN :startTime AND :endTime")
    Double getAverageVolumeInTimeRange(
            @Param("symbol") String symbol,
            @Param("startTime") LocalDateTime startTime,
            @Param("endTime") LocalDateTime endTime);

    @Query("SELECT STDDEV(md.closePrice) FROM MarketDataTs md WHERE md.symbol = :symbol " +
            "AND md.timestamp BETWEEN :startTime AND :endTime")
    BigDecimal getPriceVolatilityInTimeRange(
            @Param("symbol") String symbol,
            @Param("startTime") LocalDateTime startTime,
            @Param("endTime") LocalDateTime endTime);

    @Query("SELECT DISTINCT md.symbol FROM MarketDataTs md")
    List<String> findDistinctSymbols();

    @Query("SELECT md FROM MarketDataTs md WHERE ABS(md.priceChangePercent) >= :threshold " +
            "ORDER BY ABS(md.priceChangePercent) DESC")
    List<MarketDataTs> findHighVolatilityData(@Param("threshold") double threshold);
}