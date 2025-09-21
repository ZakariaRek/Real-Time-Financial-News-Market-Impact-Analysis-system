package com.alert_signals.AlertSignal.dto;

import com.fasterxml.jackson.annotation.JsonFormat;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class TradingSignalDto {

    private UUID id;
    private UUID predictionId;
    private String symbol;
    private String signalType;
    private String direction;
    private BigDecimal strength;
    private BigDecimal confidence;
    private String status;

    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss")
    private LocalDateTime generatedAt;

    private BigDecimal actualReturnPercent;
}