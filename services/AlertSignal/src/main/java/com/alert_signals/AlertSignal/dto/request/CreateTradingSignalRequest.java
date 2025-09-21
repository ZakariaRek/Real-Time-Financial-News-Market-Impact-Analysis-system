package com.alert_signals.AlertSignal.dto.request;

import jakarta.validation.constraints.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CreateTradingSignalRequest {

    private UUID predictionId;

    @NotBlank(message = "Symbol is required")
    @Size(max = 10, message = "Symbol must not exceed 10 characters")
    private String symbol;

    @NotBlank(message = "Signal type is required")
    private String signalType;

    @NotBlank(message = "Direction is required")
    @Pattern(regexp = "BUY|SELL|HOLD", message = "Direction must be BUY, SELL, or HOLD")
    private String direction;

    @DecimalMin(value = "0.0", message = "Strength must be non-negative")
    @DecimalMax(value = "1.0", message = "Strength must not exceed 1.0")
    private BigDecimal strength;

    @NotNull(message = "Confidence is required")
    @DecimalMin(value = "0.0", message = "Confidence must be non-negative")
    @DecimalMax(value = "1.0", message = "Confidence must not exceed 1.0")
    private BigDecimal confidence;

    @NotBlank(message = "Status is required")
    @Pattern(regexp = "ACTIVE|INACTIVE|EXECUTED|EXPIRED", message = "Invalid status")
    private String status;
}