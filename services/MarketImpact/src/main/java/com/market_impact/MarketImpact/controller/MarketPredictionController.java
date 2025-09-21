package com.market_impact.MarketImpact.controller;

import com.market_impact.MarketImpact.dto.MarketPredictionDto;
import com.market_impact.MarketImpact.Services.MarketPredictionService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.web.PageableDefault;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/api/v1/market-predictions")
@RequiredArgsConstructor
@Slf4j
public class MarketPredictionController {

    private final MarketPredictionService marketPredictionService;

    @PostMapping
    public ResponseEntity<MarketPredictionDto.Response> createPrediction(
            @Valid @RequestBody MarketPredictionDto.Request request) {
        log.info("Creating market prediction for symbol: {}", request.getSymbol());

        MarketPredictionDto.Response response = marketPredictionService.createPrediction(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @GetMapping("/{id}")
    public ResponseEntity<MarketPredictionDto.Response> getPredictionById(@PathVariable UUID id) {
        log.debug("Fetching market prediction by ID: {}", id);

        return marketPredictionService.getPredictionById(id)
                .map(prediction -> ResponseEntity.ok(prediction))
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/symbol/{symbol}")
    public ResponseEntity<List<MarketPredictionDto.Response>> getPredictionsBySymbol(
            @PathVariable String symbol) {
        log.debug("Fetching market predictions for symbol: {}", symbol);

        List<MarketPredictionDto.Response> predictions =
                marketPredictionService.getPredictionsBySymbol(symbol);
        return ResponseEntity.ok(predictions);
    }

    @GetMapping("/symbol/{symbol}/summary")
    public ResponseEntity<List<MarketPredictionDto.Summary>> getPredictionsSummaryBySymbol(
            @PathVariable String symbol) {
        log.debug("Fetching predictions summary for symbol: {}", symbol);

        List<MarketPredictionDto.Summary> summaries =
                marketPredictionService.getPredictionsSummaryBySymbol(symbol);
        return ResponseEntity.ok(summaries);
    }

    @GetMapping("/symbol/{symbol}/paged")
    public ResponseEntity<Page<MarketPredictionDto.Response>> getPredictionsBySymbolPaged(
            @PathVariable String symbol,
            @PageableDefault(size = 20, sort = "predictionTimestamp") Pageable pageable) {
        log.debug("Fetching paginated market predictions for symbol: {}", symbol);

        Page<MarketPredictionDto.Response> predictions =
                marketPredictionService.getPredictionsBySymbolPaged(symbol, pageable);
        return ResponseEntity.ok(predictions);
    }

    @GetMapping("/recent")
    public ResponseEntity<List<MarketPredictionDto.Response>> getRecentPredictions(
            @RequestParam(defaultValue = "24") int hours) {
        log.debug("Fetching recent predictions from last {} hours", hours);

        List<MarketPredictionDto.Response> predictions =
                marketPredictionService.getRecentPredictions(hours);
        return ResponseEntity.ok(predictions);
    }

    @GetMapping("/high-confidence")
    public ResponseEntity<List<MarketPredictionDto.Response>> getHighConfidencePredictions(
            @RequestParam(defaultValue = "0.8") double minConfidence) {
        log.debug("Fetching high confidence predictions with confidence >= {}", minConfidence);

        List<MarketPredictionDto.Response> predictions =
                marketPredictionService.getHighConfidencePredictions(minConfidence);
        return ResponseEntity.ok(predictions);
    }

    @GetMapping("/symbol/{symbol}/time-range")
    public ResponseEntity<List<MarketPredictionDto.Response>> getPredictionsByTimeRange(
            @PathVariable String symbol,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime startTime,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime endTime) {
        log.debug("Fetching predictions for symbol {} between {} and {}", symbol, startTime, endTime);

        List<MarketPredictionDto.Response> predictions =
                marketPredictionService.getPredictionsByTimeRange(symbol, startTime, endTime);
        return ResponseEntity.ok(predictions);
    }

    @GetMapping("/symbols")
    public ResponseEntity<List<String>> getAvailableSymbols() {
        log.debug("Fetching available symbols");

        List<String> symbols = marketPredictionService.getAvailableSymbols();
        return ResponseEntity.ok(symbols);
    }

    @GetMapping("/symbol/{symbol}/latest")
    public ResponseEntity<MarketPredictionDto.Response> getLatestPredictionForSymbol(
            @PathVariable String symbol) {
        log.debug("Fetching latest prediction for symbol: {}", symbol);

        return marketPredictionService.getLatestPredictionForSymbol(symbol)
                .map(prediction -> ResponseEntity.ok(prediction))
                .orElse(ResponseEntity.notFound().build());
    }

    @PutMapping("/{id}")
    public ResponseEntity<MarketPredictionDto.Response> updatePrediction(
            @PathVariable UUID id,
            @Valid @RequestBody MarketPredictionDto.Request request) {
        log.info("Updating market prediction with ID: {}", id);

        try {
            MarketPredictionDto.Response response =
                    marketPredictionService.updatePrediction(id, request);
            return ResponseEntity.ok(response);
        } catch (RuntimeException e) {
            log.error("Error updating market prediction: {}", e.getMessage());
            return ResponseEntity.notFound().build();
        }
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deletePrediction(@PathVariable UUID id) {
        log.info("Deleting market prediction with ID: {}", id);

        try {
            marketPredictionService.deletePrediction(id);
            return ResponseEntity.noContent().build();
        } catch (RuntimeException e) {
            log.error("Error deleting market prediction: {}", e.getMessage());
            return ResponseEntity.notFound().build();
        }
    }
}