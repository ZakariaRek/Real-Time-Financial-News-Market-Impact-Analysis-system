package com.alert_signals.AlertSignal.service;

import com.alert_signals.AlertSignal.entity.TradingSignal;
import com.alert_signals.AlertSignal.repository.TradingSignalRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
@Transactional
public class TradingSignalService {

    private final TradingSignalRepository tradingSignalRepository;

    public TradingSignal createSignal(TradingSignal signal) {
        log.info("Creating new trading signal for symbol: {}", signal.getSymbol());
        return tradingSignalRepository.save(signal);
    }

    public Optional<TradingSignal> getSignalById(UUID id) {
        return tradingSignalRepository.findById(id);
    }

    public List<TradingSignal> getSignalsBySymbol(String symbol) {
        return tradingSignalRepository.findBySymbol(symbol);
    }

    public List<TradingSignal> getSignalsByStatus(String status) {
        return tradingSignalRepository.findByStatus(status);
    }

    public List<TradingSignal> getSignalsBySymbolAndStatus(String symbol, String status) {
        return tradingSignalRepository.findBySymbolAndStatus(symbol, status);
    }

    public List<TradingSignal> getSignalsByDateRange(LocalDateTime startDate, LocalDateTime endDate) {
        return tradingSignalRepository.findByGeneratedAtBetween(startDate, endDate);
    }

    public TradingSignal updateSignal(TradingSignal signal) {
        log.info("Updating trading signal with ID: {}", signal.getId());
        return tradingSignalRepository.save(signal);
    }

    public void deleteSignal(UUID id) {
        log.info("Deleting trading signal with ID: {}", id);
        tradingSignalRepository.deleteById(id);
    }

    public List<TradingSignal> getAllSignals() {
        return tradingSignalRepository.findAll();
    }
}
