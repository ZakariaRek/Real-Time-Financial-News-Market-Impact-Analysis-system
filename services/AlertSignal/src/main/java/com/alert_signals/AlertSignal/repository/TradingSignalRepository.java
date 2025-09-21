package com.alert_signals.AlertSignal.repository;

import com.alert_signals.AlertSignal.entity.TradingSignal;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

@Repository
public interface TradingSignalRepository extends JpaRepository<TradingSignal, UUID> {

    List<TradingSignal> findBySymbol(String symbol);

    List<TradingSignal> findBySignalType(String signalType);

    List<TradingSignal> findByStatus(String status);

    List<TradingSignal> findBySymbolAndStatus(String symbol, String status);

    @Query("SELECT ts FROM TradingSignal ts WHERE ts.generatedAt BETWEEN :startDate AND :endDate")
    List<TradingSignal> findByGeneratedAtBetween(@Param("startDate") LocalDateTime startDate,
                                                 @Param("endDate") LocalDateTime endDate);

    @Query("SELECT ts FROM TradingSignal ts WHERE ts.symbol IN :symbols AND ts.status = :status")
    List<TradingSignal> findBySymbolsAndStatus(@Param("symbols") List<String> symbols,
                                               @Param("status") String status);
}
