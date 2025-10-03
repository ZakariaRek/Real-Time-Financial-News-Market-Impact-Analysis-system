package com.alert_signals.AlertSignal.controller;

import com.alert_signals.AlertSignal.dto.SignalPerformanceDto;
import com.alert_signals.AlertSignal.dto.TradingSignalDto;
import com.alert_signals.AlertSignal.dto.response.ApiResponse;
import com.alert_signals.AlertSignal.entity.SignalPerformance;
import com.alert_signals.AlertSignal.entity.TradingSignal;
import com.alert_signals.AlertSignal.mapper.SignalPerformanceMapper;
import com.alert_signals.AlertSignal.mapper.TradingSignalMapper;
import com.alert_signals.AlertSignal.service.SignalEvaluationService;
import com.alert_signals.AlertSignal.service.SignalPerformanceService;
import com.alert_signals.AlertSignal.service.TradingSignalService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

/**
 * REST Controller for signal queries
 * Implements REST API endpoints from diagram:
 * - "REST API: Get signals by symbol"
 * - "REST API: Get signal performance"
 */
@RestController
@RequestMapping("/api/signals")
@RequiredArgsConstructor
@Slf4j
@CrossOrigin(origins = "*") // Configure appropriately for production
public class SignalQueryController {

    private final TradingSignalService tradingSignalService;
    private final SignalPerformanceService signalPerformanceService;
    private final SignalEvaluationService signalEvaluationService;
    private final TradingSignalMapper tradingSignalMapper;
    private final SignalPerformanceMapper signalPerformanceMapper;

    /**
     * Get active signals (from Redis cache first, then DB)
     * Implements: "ASS->>+RSIGNAL: Get active signals"
     */
    @GetMapping("/active")
    public ResponseEntity<ApiResponse<List<TradingSignalDto>>> getActiveSignals() {
        log.info("REST: Get active signals");

        List<TradingSignal> signals = tradingSignalService.getSignalsByStatus("ACTIVE");
        List<TradingSignalDto> signalDtos = signals.stream()
                .map(tradingSignalMapper::toDto)
                .collect(Collectors.toList());

        return ResponseEntity.ok(ApiResponse.success(signalDtos));
    }

    /**
     * Get signals by symbol
     * Implements: "REST API: Get signals by symbol"
     */
    @GetMapping("/symbol/{symbol}")
    public ResponseEntity<ApiResponse<List<TradingSignalDto>>> getSignalsBySymbol(
            @PathVariable String symbol,
            @RequestParam(required = false) String status) {

        log.info("REST: Get signals for symbol: {}, status: {}", symbol, status);

        List<TradingSignal> signals;
        if (status != null) {
            // Try cache first for active signals
            if ("ACTIVE".equals(status)) {
                signals = signalEvaluationService.getActiveSignalsFromCache(symbol);
                if (signals.isEmpty()) {
                    signals = tradingSignalService.getSignalsBySymbolAndStatus(symbol, status);
                }
            } else {
                signals = tradingSignalService.getSignalsBySymbolAndStatus(symbol, status);
            }
        } else {
            signals = tradingSignalService.getSignalsBySymbol(symbol);
        }

        List<TradingSignalDto> signalDtos = signals.stream()
                .map(tradingSignalMapper::toDto)
                .collect(Collectors.toList());

        return ResponseEntity.ok(ApiResponse.success(signalDtos));
    }

    /**
     * Get signal details by ID
     * Implements: "ASS->>+PGSIGNAL: Get signal details"
     */
    @GetMapping("/{id}")
    public ResponseEntity<ApiResponse<TradingSignalDto>> getSignalById(
            @PathVariable UUID id) {

        log.info("REST: Get signal by ID: {}", id);

        return tradingSignalService.getSignalById(id)
                .map(signal -> ResponseEntity.ok(
                        ApiResponse.success(tradingSignalMapper.toDto(signal))))
                .orElse(ResponseEntity.notFound().build());
    }

    /**
     * Get signal performance
     * Implements: "REST API: Get signal performance"
     */
    @GetMapping("/performance/signal/{signalId}")
    public ResponseEntity<ApiResponse<List<SignalPerformanceDto>>> getSignalPerformance(
            @PathVariable UUID signalId) {

        log.info("REST: Get performance for signal: {}", signalId);

        List<SignalPerformance> performances = signalPerformanceService
                .getPerformanceBySignalId(signalId);

        List<SignalPerformanceDto> performanceDtos = performances.stream()
                .map(signalPerformanceMapper::toDto)
                .collect(Collectors.toList());

        return ResponseEntity.ok(ApiResponse.success(performanceDtos));
    }

    /**
     * Get latest performance for a signal
     */
    @GetMapping("/performance/signal/{signalId}/latest")
    public ResponseEntity<ApiResponse<List<SignalPerformanceDto>>> getLatestPerformance(
            @PathVariable UUID signalId,
            @RequestParam(defaultValue = "10") int limit) {

        log.info("REST: Get latest {} performance records for signal: {}", limit, signalId);

        List<SignalPerformance> performances = signalPerformanceService
                .getLatestPerformanceForSignal(signalId);

        List<SignalPerformanceDto> performanceDtos = performances.stream()
                .limit(limit)
                .map(signalPerformanceMapper::toDto)
                .collect(Collectors.toList());

        return ResponseEntity.ok(ApiResponse.success(performanceDtos));
    }

    /**
     * Get average accuracy for a signal
     */
    @GetMapping("/performance/signal/{signalId}/accuracy")
    public ResponseEntity<ApiResponse<Double>> getAverageAccuracy(
            @PathVariable UUID signalId) {

        log.info("REST: Get average accuracy for signal: {}", signalId);

        Double avgAccuracy = signalPerformanceService.getAverageAccuracyForSignal(signalId);
        return ResponseEntity.ok(ApiResponse.success(avgAccuracy));
    }

    /**
     * Get high-confidence signals (>0.8)
     */
    @GetMapping("/high-confidence")
    public ResponseEntity<ApiResponse<List<TradingSignalDto>>> getHighConfidenceSignals(
            @RequestParam(defaultValue = "0.8") double minConfidence) {

        log.info("REST: Get high-confidence signals (>= {})", minConfidence);

        // Get all active signals and filter by confidence
        List<TradingSignal> signals = tradingSignalService.getSignalsByStatus("ACTIVE")
                .stream()
                .filter(s -> s.getConfidence().doubleValue() >= minConfidence)
                .collect(Collectors.toList());

        List<TradingSignalDto> signalDtos = signals.stream()
                .map(tradingSignalMapper::toDto)
                .collect(Collectors.toList());

        return ResponseEntity.ok(ApiResponse.success(signalDtos));
    }

    /**
     * Get signals by date range
     */
    @GetMapping("/date-range")
    public ResponseEntity<ApiResponse<List<TradingSignalDto>>> getSignalsByDateRange(
            @RequestParam String startDate,
            @RequestParam String endDate) {

        log.info("REST: Get signals from {} to {}", startDate, endDate);

        List<TradingSignal> signals = tradingSignalService.getSignalsByDateRange(
                java.time.LocalDateTime.parse(startDate),
                java.time.LocalDateTime.parse(endDate)
        );

        List<TradingSignalDto> signalDtos = signals.stream()
                .map(tradingSignalMapper::toDto)
                .collect(Collectors.toList());

        return ResponseEntity.ok(ApiResponse.success(signalDtos));
    }
}