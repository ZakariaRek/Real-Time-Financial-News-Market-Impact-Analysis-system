package com.alert_signals.AlertSignal.mapper;

import com.alert_signals.AlertSignal.dto.SignalPerformanceDto;
import com.alert_signals.AlertSignal.entity.SignalPerformance;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.stream.Collectors;

@Component
public class SignalPerformanceMapper {

    public SignalPerformanceDto toDto(SignalPerformance entity) {
        if (entity == null) {
            return null;
        }

        return SignalPerformanceDto.builder()
                .id(entity.getId())
                .signalId(entity.getSignalId())
                .performanceDate(entity.getPerformanceDate())
                .return1d(entity.getReturn1d())
                .return1w(entity.getReturn1w())
                .maxDrawdown(entity.getMaxDrawdown())
                .accuracy(entity.getAccuracy())
                .build();
    }

    public SignalPerformance toEntity(SignalPerformanceDto dto) {
        if (dto == null) {
            return null;
        }

        return SignalPerformance.builder()
                .id(dto.getId())
                .signalId(dto.getSignalId())
                .performanceDate(dto.getPerformanceDate())
                .return1d(dto.getReturn1d())
                .return1w(dto.getReturn1w())
                .maxDrawdown(dto.getMaxDrawdown())
                .accuracy(dto.getAccuracy())
                .build();
    }

    public List<SignalPerformanceDto> toDtoList(List<SignalPerformance> entities) {
        if (entities == null) {
            return null;
        }

        return entities.stream()
                .map(this::toDto)
                .collect(Collectors.toList());
    }

    public List<SignalPerformance> toEntityList(List<SignalPerformanceDto> dtos) {
        if (dtos == null) {
            return null;
        }

        return dtos.stream()
                .map(this::toEntity)
                .collect(Collectors.toList());
    }
}