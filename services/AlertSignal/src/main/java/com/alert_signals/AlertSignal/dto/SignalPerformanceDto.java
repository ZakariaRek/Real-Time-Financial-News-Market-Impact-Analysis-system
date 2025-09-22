package com.alert_signals.AlertSignal.dto;

import com.fasterxml.jackson.annotation.JsonFormat;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class SignalPerformanceDto {

    private UUID id;
    private UUID signalId;

    @JsonFormat(pattern = "yyyy-MM-dd")
    private LocalDate performanceDate;

    private BigDecimal return1d;
    private BigDecimal return1w;
    private BigDecimal maxDrawdown;
    private BigDecimal accuracy;
}