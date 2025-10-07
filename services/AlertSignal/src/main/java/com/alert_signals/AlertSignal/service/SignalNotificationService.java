package com.alert_signals.AlertSignal.service;

import com.alert_signals.AlertSignal.dto.SignalNotificationDto;
import com.alert_signals.AlertSignal.entity.SignalPerformance;
import com.alert_signals.AlertSignal.entity.TradingSignal;
import com.alert_signals.AlertSignal.repository.SignalPerformanceRepository;
import com.alert_signals.AlertSignal.repository.TradingSignalRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Service
@RequiredArgsConstructor
@Slf4j
public class SignalNotificationService {

    private final TradingSignalRepository signalRepository;
    private final SignalPerformanceRepository performanceRepository;
    private final SimpMessagingTemplate messagingTemplate;

    @Async
    @Transactional
    public void processAndNotifySignal(TradingSignal signal, String triggeredRule) {
        try {
            log.info("📝 Processing signal for notification: signalId={}, symbol={}",
                    signal.getId(), signal.getSymbol());

            // Save signal
            TradingSignal savedSignal = signalRepository.save(signal);
            log.info("✅ Signal saved to database: {}", savedSignal.getId());

            // Initialize performance tracking
            SignalPerformance performance = SignalPerformance.builder()
                    .signalId(savedSignal.getId())  // ✅ Use signalId instead
                    .performanceDate(LocalDateTime.now().toLocalDate())
                    .accuracy(BigDecimal.ZERO)
                    .return1d(BigDecimal.ZERO)
                    .return1w(BigDecimal.ZERO)
                    .maxDrawdown(BigDecimal.ZERO)
                    .build();

            performanceRepository.save(performance);
            log.info("Initialized performance tracking for signal: {}", savedSignal.getId());

            // ✅ BROADCAST TO WEBSOCKET
            broadcastSignalToWebSocket(savedSignal, triggeredRule);

            log.info("✅ Successfully processed signal: {}", savedSignal.getId());

        } catch (Exception e) {
            log.error("❌ Error processing signal notification: {}", e.getMessage(), e);
        }
    }

    private void broadcastSignalToWebSocket(TradingSignal signal, String triggeredRule) {
        try {
            log.info("📤 Broadcasting signal to WebSocket subscribers: signalId={}, symbol={}",
                    signal.getId(), signal.getSymbol());

            SignalNotificationDto notification = SignalNotificationDto.builder()
                    .signalId(signal.getId().toString())
                    .predictionId(signal.getPredictionId().toString())
                    .symbol(signal.getSymbol())
                    .signalType(signal.getSignalType())
                    .direction(signal.getDirection())
                    .confidence(signal.getConfidence())
                    .strength(signal.getStrength())
                    .triggeredRule(triggeredRule)
                    .timestamp(signal.getGeneratedAt())
                    .build();

            // Send to WebSocket topic
            messagingTemplate.convertAndSend("/topic/signals", notification);

            log.info("✅ Signal broadcast to WebSocket successfully: signalId={}", signal.getId());
            log.debug("📨 Broadcast payload: {}", notification);

        } catch (Exception e) {
            log.error("❌ Failed to broadcast signal to WebSocket: {}", e.getMessage(), e);
            e.printStackTrace();
        }
    }

    /**
     * Broadcast test signal for debugging
     */
    public void broadcastTestSignal() {
        try {
            log.info("🧪 Broadcasting test signal to WebSocket");

            SignalNotificationDto testNotification = SignalNotificationDto.builder()
                    .signalId("test-" + System.currentTimeMillis())
                    .predictionId("test-prediction")
                    .symbol("TEST")
                    .signalType("TEST_SIGNAL")
                    .direction("UP")
                    .confidence(BigDecimal.valueOf(0.99))
                    .strength(BigDecimal.valueOf(100.0))
                    .triggeredRule("Test Rule")
                    .timestamp(LocalDateTime.now())
                    .build();

            messagingTemplate.convertAndSend("/topic/signals", testNotification);

            log.info("✅ Test signal broadcast successfully");

        } catch (Exception e) {
            log.error("❌ Failed to broadcast test signal: {}", e.getMessage(), e);
        }
    }
}