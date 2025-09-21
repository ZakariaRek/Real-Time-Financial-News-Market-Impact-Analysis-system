package com.alert_signals.AlertSignal.service;

import com.alert_signals.AlertSignal.entity.SignalPerformance;
import com.alert_signals.AlertSignal.repository.SignalPerformanceRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
@Transactional
public class SignalPerformanceService {

    private final SignalPerformanceRepository signalPerformanceRepository;

    public SignalPerformance createPerformanceRecord(SignalPerformance performance) {
        log.info("Creating performance record for signal ID: {}", performance.getSignalId());
        return signalPerformanceRepository.save(performance);
    }

    public Optional<SignalPerformance> getPerformanceById(UUID id) {
        return signalPerformanceRepository.findById(id);
    }

    public List<SignalPerformance> getPerformanceBySignalId(UUID signalId) {
        return signalPerformanceRepository.findBySignalId(signalId);
    }

    public Optional<SignalPerformance> getPerformanceBySignalAndDate(UUID signalId, LocalDate date) {
        return signalPerformanceRepository.findBySignalIdAndPerformanceDate(signalId, date);
    }

    public List<SignalPerformance> getPerformanceByDateRange(LocalDate startDate, LocalDate endDate) {
        return signalPerformanceRepository.findByPerformanceDateBetween(startDate, endDate);
    }

    public Double getAverageAccuracyForSignal(UUID signalId) {
        return signalPerformanceRepository.getAverageAccuracyBySignalId(signalId);
    }

    public List<SignalPerformance> getLatestPerformanceForSignal(UUID signalId) {
        return signalPerformanceRepository.findBySignalIdOrderByPerformanceDateDesc(signalId);
    }

    public SignalPerformance updatePerformance(SignalPerformance performance) {
        log.info("Updating performance record with ID: {}", performance.getId());
        return signalPerformanceRepository.save(performance);
    }

    public void deletePerformance(UUID id) {
        log.info("Deleting performance record with ID: {}", id);
        signalPerformanceRepository.deleteById(id);
    }
}