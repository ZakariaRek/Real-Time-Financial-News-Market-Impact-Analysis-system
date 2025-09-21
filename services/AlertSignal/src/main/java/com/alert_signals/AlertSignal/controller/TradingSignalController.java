package com.alert_signals.AlertSignal.controller;

import com.alert_signals.AlertSignal.dto.TradingSignalDto;
import com.alert_signals.AlertSignal.entity.TradingSignal;
import com.alert_signals.AlertSignal.mapper.TradingSignalMapper;
import com.alert_signals.AlertSignal.service.TradingSignalService;
import lombok.RequiredArgsConstructor;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@RestController
@RequestMapping("/trading-signals")
@RequiredArgsConstructor
public class TradingSignalController {

    private final TradingSignalService tradingSignalService;
    private final TradingSignalMapper tradingSignalMapper;

    @PostMapping
    public ResponseEntity<TradingSignalDto> createSignal(@RequestBody TradingSignalDto signalDto) {
        TradingSignal signal = tradingSignalMapper.toEntity(signalDto);
        TradingSignal savedSignal = tradingSignalService.createSignal(signal);
        TradingSignalDto responseDto = tradingSignalMapper.toDto(savedSignal);
        return new ResponseEntity<>(responseDto, HttpStatus.CREATED);
    }

    @GetMapping("/{id}")
    public ResponseEntity<TradingSignalDto> getSignalById(@PathVariable UUID id) {
        Optional<TradingSignal> signal = tradingSignalService.getSignalById(id);
        return signal.map(s -> ResponseEntity.ok(tradingSignalMapper.toDto(s)))
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping
    public ResponseEntity<List<TradingSignalDto>> getAllSignals() {
        List<TradingSignal> signals = tradingSignalService.getAllSignals();
        List<TradingSignalDto> signalDtos = tradingSignalMapper.toDtoList(signals);
        return ResponseEntity.ok(signalDtos);
    }

    @GetMapping("/symbol/{symbol}")
    public ResponseEntity<List<TradingSignalDto>> getSignalsBySymbol(@PathVariable String symbol) {
        List<TradingSignal> signals = tradingSignalService.getSignalsBySymbol(symbol);
        List<TradingSignalDto> signalDtos = tradingSignalMapper.toDtoList(signals);
        return ResponseEntity.ok(signalDtos);
    }

    @GetMapping("/status/{status}")
    public ResponseEntity<List<TradingSignalDto>> getSignalsByStatus(@PathVariable String status) {
        List<TradingSignal> signals = tradingSignalService.getSignalsByStatus(status);
        List<TradingSignalDto> signalDtos = tradingSignalMapper.toDtoList(signals);
        return ResponseEntity.ok(signalDtos);
    }

    @GetMapping("/symbol/{symbol}/status/{status}")
    public ResponseEntity<List<TradingSignalDto>> getSignalsBySymbolAndStatus(
            @PathVariable String symbol,
            @PathVariable String status) {
        List<TradingSignal> signals = tradingSignalService.getSignalsBySymbolAndStatus(symbol, status);
        List<TradingSignalDto> signalDtos = tradingSignalMapper.toDtoList(signals);
        return ResponseEntity.ok(signalDtos);
    }

    @GetMapping("/date-range")
    public ResponseEntity<List<TradingSignalDto>> getSignalsByDateRange(
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime startDate,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime endDate) {
        List<TradingSignal> signals = tradingSignalService.getSignalsByDateRange(startDate, endDate);
        List<TradingSignalDto> signalDtos = tradingSignalMapper.toDtoList(signals);
        return ResponseEntity.ok(signalDtos);
    }

    @PutMapping("/{id}")
    public ResponseEntity<TradingSignalDto> updateSignal(
            @PathVariable UUID id,
            @RequestBody TradingSignalDto signalDto) {
        Optional<TradingSignal> existingSignal = tradingSignalService.getSignalById(id);
        if (existingSignal.isEmpty()) {
            return ResponseEntity.notFound().build();
        }

        signalDto.setId(id);
        TradingSignal signal = tradingSignalMapper.toEntity(signalDto);
        TradingSignal updatedSignal = tradingSignalService.updateSignal(signal);
        TradingSignalDto responseDto = tradingSignalMapper.toDto(updatedSignal);
        return ResponseEntity.ok(responseDto);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteSignal(@PathVariable UUID id) {
        Optional<TradingSignal> existingSignal = tradingSignalService.getSignalById(id);
        if (existingSignal.isEmpty()) {
            return ResponseEntity.notFound().build();
        }

        tradingSignalService.deleteSignal(id);
        return ResponseEntity.noContent().build();
    }
}