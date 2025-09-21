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
public class UserSubscriptionDto {

    private UUID id;
    private String userId;
    private String subscriptionName;
    private List<String> symbols;
    private List<String> signalTypes;
    private BigDecimal minConfidence;
    private List<String> deliveryMethods;
    private Boolean isActive;
}