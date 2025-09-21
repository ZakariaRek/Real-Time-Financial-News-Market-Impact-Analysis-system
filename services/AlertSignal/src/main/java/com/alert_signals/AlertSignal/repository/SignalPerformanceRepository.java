package com.alert_signals.AlertSignal.repository;

import com.alert_signals.AlertSignal.entity.SignalPerformance;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface SignalPerformanceRepository extends JpaRepository<SignalPerformance, UUID> {

    List<SignalPerformance> findBySignalId(UUID signalId);

    Optional<SignalPerformance> findBySignalIdAndPerformanceDate(UUID signalId, LocalDate performanceDate);

    @Query("SELECT sp FROM SignalPerformance sp WHERE sp.performanceDate BETWEEN :startDate AND :endDate")
    List<SignalPerformance> findByPerformanceDateBetween(@Param("startDate") LocalDate startDate,
                                                         @Param("endDate") LocalDate endDate);

    @Query("SELECT AVG(sp.accuracy) FROM SignalPerformance sp WHERE sp.signalId = :signalId")
    Double getAverageAccuracyBySignalId(@Param("signalId") UUID signalId);

    @Query("SELECT sp FROM SignalPerformance sp WHERE sp.signalId = :signalId ORDER BY sp.performanceDate DESC")
    List<SignalPerformance> findBySignalIdOrderByPerformanceDateDesc(@Param("signalId") UUID signalId);
}
