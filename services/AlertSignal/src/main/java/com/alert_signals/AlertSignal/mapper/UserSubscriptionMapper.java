package com.alert_signals.AlertSignal.mapper;

import com.alert_signals.AlertSignal.dto.UserSubscriptionDto;
import com.alert_signals.AlertSignal.entity.UserSubscription;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.stream.Collectors;

@Component
public class UserSubscriptionMapper {

    public UserSubscriptionDto toDto(UserSubscription entity) {
        if (entity == null) {
            return null;
        }

        return UserSubscriptionDto.builder()
                .id(entity.getId())
                .userId(entity.getUserId())
                .subscriptionName(entity.getSubscriptionName())
                .symbols(entity.getSymbols())
                .signalTypes(entity.getSignalTypes())
                .minConfidence(entity.getMinConfidence())
                .deliveryMethods(entity.getDeliveryMethods())
                .isActive(entity.getIsActive())
                .build();
    }

    public UserSubscription toEntity(UserSubscriptionDto dto) {
        if (dto == null) {
            return null;
        }

        return UserSubscription.builder()
                .id(dto.getId())
                .userId(dto.getUserId())
                .subscriptionName(dto.getSubscriptionName())
                .symbols(dto.getSymbols())
                .signalTypes(dto.getSignalTypes())
                .minConfidence(dto.getMinConfidence())
                .deliveryMethods(dto.getDeliveryMethods())
                .isActive(dto.getIsActive())
                .build();
    }

    public List<UserSubscriptionDto> toDtoList(List<UserSubscription> entities) {
        if (entities == null) {
            return null;
        }

        return entities.stream()
                .map(this::toDto)
                .collect(Collectors.toList());
    }

    public List<UserSubscription> toEntityList(List<UserSubscriptionDto> dtos) {
        if (dtos == null) {
            return null;
        }

        return dtos.stream()
                .map(this::toEntity)
                .collect(Collectors.toList());
    }
}