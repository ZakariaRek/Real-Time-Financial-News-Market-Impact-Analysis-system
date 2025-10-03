package com.alert_signals.AlertSignal.controller;

import com.alert_signals.AlertSignal.dto.TradingSignalDto;
import com.alert_signals.AlertSignal.entity.TradingSignal;
import com.alert_signals.AlertSignal.mapper.TradingSignalMapper;
import com.alert_signals.AlertSignal.service.SignalEvaluationService;
import com.alert_signals.AlertSignal.service.TradingSignalService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.SendTo;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Controller;

import java.util.List;
import java.util.stream.Collectors;

/**
 * WebSocket controller for real-time signal updates
 * Implements real-time push from diagram: "WEB->>+ASS: WebSocket connection"
 */
@Controller
@RequiredArgsConstructor
@Slf4j
public class SignalWebSocketController {

    private final SimpMessagingTemplate messagingTemplate;
    private final TradingSignalService tradingSignalService;
    private final SignalEvaluationService signalEvaluationService;
    private final TradingSignalMapper tradingSignalMapper;

    /**
     * Handle subscription to signal updates for a specific symbol
     */
    @MessageMapping("/signals/subscribe")
    @SendTo("/topic/signals")
    public List<TradingSignalDto> subscribeToSignals(String symbol) {
        log.info("Client subscribed to signals for symbol: {}", symbol);

        // Get active signals from cache first
        List<TradingSignal> activeSignals = signalEvaluationService
                .getActiveSignalsFromCache(symbol);

        if (activeSignals.isEmpty()) {
            // Fallback to database
            activeSignals = tradingSignalService.getSignalsBySymbolAndStatus(
                    symbol, "ACTIVE");
        }

        return activeSignals.stream()
                .map(tradingSignalMapper::toDto)
                .collect(Collectors.toList());
    }

    /**
     * Broadcast new signal to all connected clients
     * Called when a new high-confidence signal is created
     */
    public void broadcastNewSignal(TradingSignal signal) {
        try {
            TradingSignalDto signalDto = tradingSignalMapper.toDto(signal);

            // Broadcast to all subscribers
            messagingTemplate.convertAndSend("/topic/signals", signalDto);

            // Send to symbol-specific topic
            messagingTemplate.convertAndSend(
                    "/topic/signals/" + signal.getSymbol(),
                    signalDto
            );

            log.info("Broadcasted signal update: {} for symbol: {}",
                    signal.getId(), signal.getSymbol());
        } catch (Exception e) {
            log.error("Failed to broadcast signal: {}", e.getMessage());
        }
    }

    /**
     * Push active signals periodically (every 5 seconds)
     * Implements the "Push signal updates" loop from diagram
     */
    @Scheduled(fixedDelay = 5000)
    public void pushActiveSignals() {
        try {
            // Get all active signals
            List<TradingSignal> activeSignals = tradingSignalService
                    .getSignalsByStatus("ACTIVE");

            if (!activeSignals.isEmpty()) {
                List<TradingSignalDto> signalDtos = activeSignals.stream()
                        .map(tradingSignalMapper::toDto)
                        .collect(Collectors.toList());

                // Broadcast to all connected clients
                messagingTemplate.convertAndSend("/topic/signals/updates", signalDtos);
            }
        } catch (Exception e) {
            log.error("Failed to push active signals: {}", e.getMessage());
        }
    }

    /**
     * Handle request for signal details
     */
    @MessageMapping("/signals/details")
    @SendTo("/queue/signal-details")
    public TradingSignalDto getSignalDetails(String signalId) {
        try {
            return tradingSignalService.getSignalById(java.util.UUID.fromString(signalId))
                    .map(tradingSignalMapper::toDto)
                    .orElse(null);
        } catch (Exception e) {
            log.error("Failed to get signal details: {}", e.getMessage());
            return null;
        }
    }
}