package com.alert_signals.AlertSignal.controller;

import com.alert_signals.AlertSignal.dto.AlertDeliveryLogDto;
import com.alert_signals.AlertSignal.entity.AlertDeliveryLog;
import com.alert_signals.AlertSignal.mapper.AlertDeliveryLogMapper;
import com.alert_signals.AlertSignal.service.AlertDeliveryService;
import lombok.RequiredArgsConstructor;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@RestController
@RequestMapping("/alert-delivery-logs")
@RequiredArgsConstructor
public class AlertDeliveryLogController {

    private final AlertDeliveryService alertDeliveryService;
    private final AlertDeliveryLogMapper alertDeliveryLogMapper;

    @PostMapping
    public ResponseEntity<AlertDeliveryLogDto> logDelivery(@RequestBody AlertDeliveryLogDto deliveryLogDto) {
        AlertDeliveryLog deliveryLog = alertDeliveryLogMapper.toEntity(deliveryLogDto);
        AlertDeliveryLog savedDeliveryLog = alertDeliveryService.logDelivery(deliveryLog);
        AlertDeliveryLogDto responseDto = alertDeliveryLogMapper.toDto(savedDeliveryLog);
        return new ResponseEntity<>(responseDto, HttpStatus.CREATED);
    }

    @GetMapping("/{id}")
    public ResponseEntity<AlertDeliveryLogDto> getDeliveryLogById(@PathVariable UUID id) {
        Optional<AlertDeliveryLog> deliveryLog = alertDeliveryService.getDeliveryLogById(id);
        return deliveryLog.map(log -> ResponseEntity.ok(alertDeliveryLogMapper.toDto(log)))
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/signal/{signalId}")
    public ResponseEntity<List<AlertDeliveryLogDto>> getDeliveryLogsBySignalId(@PathVariable UUID signalId) {
        List<AlertDeliveryLog> deliveryLogs = alertDeliveryService.getDeliveryLogsBySignalId(signalId);
        List<AlertDeliveryLogDto> deliveryLogDtos = alertDeliveryLogMapper.toDtoList(deliveryLogs);
        return ResponseEntity.ok(deliveryLogDtos);
    }

    @GetMapping("/subscription/{subscriptionId}")
    public ResponseEntity<List<AlertDeliveryLogDto>> getDeliveryLogsBySubscriptionId(@PathVariable UUID subscriptionId) {
        List<AlertDeliveryLog> deliveryLogs = alertDeliveryService.getDeliveryLogsBySubscriptionId(subscriptionId);
        List<AlertDeliveryLogDto> deliveryLogDtos = alertDeliveryLogMapper.toDtoList(deliveryLogs);
        return ResponseEntity.ok(deliveryLogDtos);
    }

    @GetMapping("/status/{status}")
    public ResponseEntity<List<AlertDeliveryLogDto>> getDeliveryLogsByStatus(@PathVariable String status) {
        List<AlertDeliveryLog> deliveryLogs = alertDeliveryService.getDeliveryLogsByStatus(status);
        List<AlertDeliveryLogDto> deliveryLogDtos = alertDeliveryLogMapper.toDtoList(deliveryLogs);
        return ResponseEntity.ok(deliveryLogDtos);
    }

    @GetMapping("/method/{deliveryMethod}")
    public ResponseEntity<List<AlertDeliveryLogDto>> getDeliveryLogsByMethod(@PathVariable String deliveryMethod) {
        List<AlertDeliveryLog> deliveryLogs = alertDeliveryService.getDeliveryLogsByMethod(deliveryMethod);
        List<AlertDeliveryLogDto> deliveryLogDtos = alertDeliveryLogMapper.toDtoList(deliveryLogs);
        return ResponseEntity.ok(deliveryLogDtos);
    }

    @GetMapping("/date-range")
    public ResponseEntity<List<AlertDeliveryLogDto>> getDeliveryLogsByDateRange(
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime startDate,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime endDate) {
        List<AlertDeliveryLog> deliveryLogs = alertDeliveryService.getDeliveryLogsByDateRange(startDate, endDate);
        List<AlertDeliveryLogDto> deliveryLogDtos = alertDeliveryLogMapper.toDtoList(deliveryLogs);
        return ResponseEntity.ok(deliveryLogDtos);
    }

    @GetMapping("/method/{deliveryMethod}/average-latency")
    public ResponseEntity<Double> getAverageDeliveryLatency(@PathVariable String deliveryMethod) {
        Double averageLatency = alertDeliveryService.getAverageDeliveryLatency(deliveryMethod);
        return ResponseEntity.ok(averageLatency);
    }

    @GetMapping("/subscription/{subscriptionId}/success-count")
    public ResponseEntity<Long> getSuccessfulDeliveryCount(@PathVariable UUID subscriptionId) {
        Long successCount = alertDeliveryService.getSuccessfulDeliveryCount(subscriptionId);
        return ResponseEntity.ok(successCount);
    }

    @PutMapping("/{id}")
    public ResponseEntity<AlertDeliveryLogDto> updateDeliveryLog(
            @PathVariable UUID id,
            @RequestBody AlertDeliveryLogDto deliveryLogDto) {
        Optional<AlertDeliveryLog> existingDeliveryLog = alertDeliveryService.getDeliveryLogById(id);
        if (existingDeliveryLog.isEmpty()) {
            return ResponseEntity.notFound().build();
        }

        deliveryLogDto.setId(id);
        AlertDeliveryLog deliveryLog = alertDeliveryLogMapper.toEntity(deliveryLogDto);
        AlertDeliveryLog updatedDeliveryLog = alertDeliveryService.updateDeliveryLog(deliveryLog);
        AlertDeliveryLogDto responseDto = alertDeliveryLogMapper.toDto(updatedDeliveryLog);
        return ResponseEntity.ok(responseDto);
    }
}