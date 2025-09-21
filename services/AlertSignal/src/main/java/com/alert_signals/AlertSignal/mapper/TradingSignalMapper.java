package com.alert_signals.AlertSignal.mapper;

import com.alert_signals.AlertSignal.dto.TradingSignalDto;
import com.alert_signals.AlertSignal.entity.TradingSignal;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.stream.Collectors;

@Component
public class TradingSignalMapper {

    public TradingSignalDto toDto(TradingSignal entity) {
        if (entity == null) {
            return null;
        }

        return TradingSignalDto.builder()
                .id(entity.getId())
                .predictionId(entity.getPredictionId())
                .symbol(entity.getSymbol())
                .signalType(entity.getSignalType())
                .direction(entity.getDirection())
                .strength(entity.getStrength())
                .confidence(entity.getConfidence())
                .status(entity.getStatus())
                .generatedAt(entity.getGeneratedAt())
                .actualReturnPercent(entity.getActualReturnPercent())
                .build();
    }

    public TradingSignal toEntity(TradingSignalDto dto) {
        if (dto == null) {
            return null;
        }

        return TradingSignal.builder()
                .id(dto.getId())
                .predictionId(dto.getPredictionId())
                .symbol(dto.getSymbol())
                .signalType(dto.getSignalType())
                .direction(dto.getDirection())
                .strength(dto.getStrength())
                .confidence(dto.getConfidence())
                .status(dto.getStatus())
                .generatedAt(dto.getGeneratedAt())
                .actualReturnPercent(dto.getActualReturnPercent())
                .build();
    }

    public List<TradingSignalDto> toDtoList(List<TradingSignal> entities) {
        if (entities == null) {
            return null;
        }

        return entities.stream()
                .map(this::toDto)
                .collect(Collectors.toList());
    }

    public List<TradingSignal> toEntityList(List<TradingSignalDto> dtos) {
        if (dtos == null) {
            return null;
        }

        return dtos.stream()
                .map(this::toEntity)
                .collect(Collectors.toList());
    }
}