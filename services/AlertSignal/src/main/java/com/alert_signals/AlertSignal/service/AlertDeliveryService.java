package com.alert_signals.AlertSignal.service;

import com.alert_signals.AlertSignal.entity.AlertDeliveryLog;
import com.alert_signals.AlertSignal.repository.AlertDeliveryLogRepository;
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
public class AlertDeliveryService {

    private final AlertDeliveryLogRepository alertDeliveryLogRepository;

    public AlertDeliveryLog logDelivery(AlertDeliveryLog deliveryLog) {
        log.info("Logging alert delivery for signal ID: {} via method: {}",
                deliveryLog.getSignalId(), deliveryLog.getDeliveryMethod());
        return alertDeliveryLogRepository.save(deliveryLog);
    }

    public Optional<AlertDeliveryLog> getDeliveryLogById(UUID id) {
        return alertDeliveryLogRepository.findById(id);
    }

    public List<AlertDeliveryLog> getDeliveryLogsBySignalId(UUID signalId) {
        return alertDeliveryLogRepository.findBySignalId(signalId);
    }

    public List<AlertDeliveryLog> getDeliveryLogsBySubscriptionId(UUID subscriptionId) {
        return alertDeliveryLogRepository.findBySubscriptionId(subscriptionId);
    }

    public List<AlertDeliveryLog> getDeliveryLogsByStatus(String status) {
        return alertDeliveryLogRepository.findByStatus(status);
    }

    public List<AlertDeliveryLog> getDeliveryLogsByMethod(String deliveryMethod) {
        return alertDeliveryLogRepository.findByDeliveryMethod(deliveryMethod);
    }

    public List<AlertDeliveryLog> getDeliveryLogsByDateRange(LocalDateTime startDate, LocalDateTime endDate) {
        return alertDeliveryLogRepository.findBySentAtBetween(startDate, endDate);
    }

    public Double getAverageDeliveryLatency(String deliveryMethod) {
        return alertDeliveryLogRepository.getAverageDeliveryLatencyByMethod(deliveryMethod);
    }

    public Long getSuccessfulDeliveryCount(UUID subscriptionId) {
        return alertDeliveryLogRepository.countSuccessfulDeliveriesBySubscription(subscriptionId);
    }

    public AlertDeliveryLog updateDeliveryLog(AlertDeliveryLog deliveryLog) {
        log.info("Updating delivery log with ID: {}", deliveryLog.getId());
        return alertDeliveryLogRepository.save(deliveryLog);
    }
}
