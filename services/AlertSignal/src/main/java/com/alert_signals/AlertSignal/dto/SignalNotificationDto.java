package com.alert_signals.AlertSignal.dto;

import com.fasterxml.jackson.annotation.JsonFormat;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class SignalNotificationDto {
    private String signalId;
    private String predictionId;
    private String symbol;
    private String signalType;
    private String direction;
    private BigDecimal confidence;
    private BigDecimal strength;
    private String triggeredRule;

    @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss")
    private LocalDateTime timestamp;
}