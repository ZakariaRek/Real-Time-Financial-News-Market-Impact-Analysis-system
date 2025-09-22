package com.alert_signals.AlertSignal.dto;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.util.List;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class SignalRulesDto {

    private UUID id;
    private String ruleName;
    private String ruleType;
    private String conditions;
    private List<String> symbols;
    private BigDecimal successRate;
    private Boolean isActive;
}