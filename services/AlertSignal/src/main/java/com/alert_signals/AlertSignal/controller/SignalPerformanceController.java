package com.alert_signals.AlertSignal.controller;

import com.alert_signals.AlertSignal.dto.SignalPerformanceDto;
import com.alert_signals.AlertSignal.entity.SignalPerformance;
import com.alert_signals.AlertSignal.mapper.SignalPerformanceMapper;
import com.alert_signals.AlertSignal.service.SignalPerformanceService;
import lombok.RequiredArgsConstructor;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@RestController
@RequestMapping("/signal-performance")
@RequiredArgsConstructor
public class SignalPerformanceController {

    private final SignalPerformanceService signalPerformanceService;
    private final SignalPerformanceMapper signalPerformanceMapper;

    @PostMapping
    public ResponseEntity<SignalPerformanceDto> createPerformanceRecord(@RequestBody SignalPerformanceDto performanceDto) {
        SignalPerformance performance = signalPerformanceMapper.toEntity(performanceDto);
        SignalPerformance savedPerformance = signalPerformanceService.createPerformanceRecord(performance);
        SignalPerformanceDto responseDto = signalPerformanceMapper.toDto(savedPerformance);
        return new ResponseEntity<>(responseDto, HttpStatus.CREATED);
    }

    @GetMapping("/{id}")
    public ResponseEntity<SignalPerformanceDto> getPerformanceById(@PathVariable UUID id) {
        Optional<SignalPerformance> performance = signalPerformanceService.getPerformanceById(id);
        return performance.map(p -> ResponseEntity.ok(signalPerformanceMapper.toDto(p)))
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/signal/{signalId}")
    public ResponseEntity<List<SignalPerformanceDto>> getPerformanceBySignalId(@PathVariable UUID signalId) {
        List<SignalPerformance> performances = signalPerformanceService.getPerformanceBySignalId(signalId);
        List<SignalPerformanceDto> performanceDtos = signalPerformanceMapper.toDtoList(performances);
        return ResponseEntity.ok(performanceDtos);
    }

    @GetMapping("/signal/{signalId}/date/{date}")
    public ResponseEntity<SignalPerformanceDto> getPerformanceBySignalAndDate(
            @PathVariable UUID signalId,
            @PathVariable @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate date) {
        Optional<SignalPerformance> performance = signalPerformanceService.getPerformanceBySignalAndDate(signalId, date);
        return performance.map(p -> ResponseEntity.ok(signalPerformanceMapper.toDto(p)))
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/date-range")
    public ResponseEntity<List<SignalPerformanceDto>> getPerformanceByDateRange(
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate startDate,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate endDate) {
        List<SignalPerformance> performances = signalPerformanceService.getPerformanceByDateRange(startDate, endDate);
        List<SignalPerformanceDto> performanceDtos = signalPerformanceMapper.toDtoList(performances);
        return ResponseEntity.ok(performanceDtos);
    }

    @GetMapping("/signal/{signalId}/average-accuracy")
    public ResponseEntity<Double> getAverageAccuracyForSignal(@PathVariable UUID signalId) {
        Double averageAccuracy = signalPerformanceService.getAverageAccuracyForSignal(signalId);
        return ResponseEntity.ok(averageAccuracy);
    }

    @GetMapping("/signal/{signalId}/latest")
    public ResponseEntity<List<SignalPerformanceDto>> getLatestPerformanceForSignal(@PathVariable UUID signalId) {
        List<SignalPerformance> performances = signalPerformanceService.getLatestPerformanceForSignal(signalId);
        List<SignalPerformanceDto> performanceDtos = signalPerformanceMapper.toDtoList(performances);
        return ResponseEntity.ok(performanceDtos);
    }

    @PutMapping("/{id}")
    public ResponseEntity<SignalPerformanceDto> updatePerformance(
            @PathVariable UUID id,
            @RequestBody SignalPerformanceDto performanceDto) {
        Optional<SignalPerformance> existingPerformance = signalPerformanceService.getPerformanceById(id);
        if (existingPerformance.isEmpty()) {
            return ResponseEntity.notFound().build();
        }

        performanceDto.setId(id);
        SignalPerformance performance = signalPerformanceMapper.toEntity(performanceDto);
        SignalPerformance updatedPerformance = signalPerformanceService.updatePerformance(performance);
        SignalPerformanceDto responseDto = signalPerformanceMapper.toDto(updatedPerformance);
        return ResponseEntity.ok(responseDto);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deletePerformance(@PathVariable UUID id) {
        Optional<SignalPerformance> existingPerformance = signalPerformanceService.getPerformanceById(id);
        if (existingPerformance.isEmpty()) {
            return ResponseEntity.notFound().build();
        }

        signalPerformanceService.deletePerformance(id);
        return ResponseEntity.noContent().build();
    }
}