package com.alert_signals.AlertSignal.dto.request;

import jakarta.validation.constraints.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.util.List;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CreateUserSubscriptionRequest {

    @NotBlank(message = "User ID is required")
    private String userId;

    @NotBlank(message = "Subscription name is required")
    @Size(max = 100, message = "Subscription name must not exceed 100 characters")
    private String subscriptionName;

    @NotEmpty(message = "At least one symbol is required")
    private List<@NotBlank(message = "Symbol cannot be blank") String> symbols;

    @NotEmpty(message = "At least one signal type is required")
    private List<@NotBlank(message = "Signal type cannot be blank") String> signalTypes;

    @DecimalMin(value = "0.0", message = "Minimum confidence must be non-negative")
    @DecimalMax(value = "1.0", message = "Minimum confidence must not exceed 1.0")
    private BigDecimal minConfidence;

    @NotEmpty(message = "At least one delivery method is required")
    private List<@NotBlank(message = "Delivery method cannot be blank") String> deliveryMethods;

    @NotNull(message = "Active status is required")
    private Boolean isActive;
}