package com.alert_signals.AlertSignal.mapper;

import com.alert_signals.AlertSignal.dto.AlertDeliveryLogDto;
import com.alert_signals.AlertSignal.entity.AlertDeliveryLog;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.stream.Collectors;

@Component
public class AlertDeliveryLogMapper {

    public AlertDeliveryLogDto toDto(AlertDeliveryLog entity) {
        if (entity == null) {
            return null;
        }

        return AlertDeliveryLogDto.builder()
                .id(entity.getId())
                .signalId(entity.getSignalId())
                .subscriptionId(entity.getSubscriptionId())
                .deliveryMethod(entity.getDeliveryMethod())
                .status(entity.getStatus())
                .sentAt(entity.getSentAt())
                .deliveryLatencyMs(entity.getDeliveryLatencyMs())
                .build();
    }

    public AlertDeliveryLog toEntity(AlertDeliveryLogDto dto) {
        if (dto == null) {
            return null;
        }

        return AlertDeliveryLog.builder()
                .id(dto.getId())
                .signalId(dto.getSignalId())
                .subscriptionId(dto.getSubscriptionId())
                .deliveryMethod(dto.getDeliveryMethod())
                .status(dto.getStatus())
                .sentAt(dto.getSentAt())
                .deliveryLatencyMs(dto.getDeliveryLatencyMs())
                .build();
    }

    public List<AlertDeliveryLogDto> toDtoList(List<AlertDeliveryLog> entities) {
        if (entities == null) {
            return null;
        }

        return entities.stream()
                .map(this::toDto)
                .collect(Collectors.toList());
    }

    public List<AlertDeliveryLog> toEntityList(List<AlertDeliveryLogDto> dtos) {
        if (dtos == null) {
            return null;
        }

        return dtos.stream()
                .map(this::toEntity)
                .collect(Collectors.toList());
    }
}