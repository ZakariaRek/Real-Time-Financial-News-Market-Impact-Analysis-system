package com.alert_signals.AlertSignal.mapper;

import com.alert_signals.AlertSignal.dto.SignalRulesDto;
import com.alert_signals.AlertSignal.entity.SignalRules;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.stream.Collectors;

@Component
public class SignalRulesMapper {

    public SignalRulesDto toDto(SignalRules entity) {
        if (entity == null) {
            return null;
        }

        return SignalRulesDto.builder()
                .id(entity.getId())
                .ruleName(entity.getRuleName())
                .ruleType(entity.getRuleType())
                .conditions(entity.getConditions())
                .symbols(entity.getSymbols())
                .successRate(entity.getSuccessRate())
                .isActive(entity.getIsActive())
                .build();
    }

    public SignalRules toEntity(SignalRulesDto dto) {
        if (dto == null) {
            return null;
        }

        return SignalRules.builder()
                .id(dto.getId())
                .ruleName(dto.getRuleName())
                .ruleType(dto.getRuleType())
                .conditions(dto.getConditions())
                .symbols(dto.getSymbols())
                .successRate(dto.getSuccessRate())
                .isActive(dto.getIsActive())
                .build();
    }

    public List<SignalRulesDto> toDtoList(List<SignalRules> entities) {
        if (entities == null) {
            return null;
        }

        return entities.stream()
                .map(this::toDto)
                .collect(Collectors.toList());
    }

    public List<SignalRules> toEntityList(List<SignalRulesDto> dtos) {
        if (dtos == null) {
            return null;
        }

        return dtos.stream()
                .map(this::toEntity)
                .collect(Collectors.toList());
    }
}