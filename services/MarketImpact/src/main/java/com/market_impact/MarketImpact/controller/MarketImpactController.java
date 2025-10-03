package com.market_impact.MarketImpact.controller;

import com.market_impact.MarketImpact.Services.MarketImpactPredictionService;
import com.market_impact.MarketImpact.entity.MarketPrediction;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/api/v1/market-impact")
@RequiredArgsConstructor
@Slf4j
public class MarketImpactController {

    private  MarketImpactPredictionService marketImpactService;

    @PostMapping("/predict")
    public ResponseEntity<MarketPrediction> generatePrediction(
            @RequestParam UUID articleId,
            @RequestParam String symbol) {

        log.info("REST: Generate prediction for article: {} and symbol: {}", articleId, symbol);

        try {
            MarketPrediction prediction = marketImpactService.generatePredictionFromSentiment(articleId, symbol);
            return ResponseEntity.ok(prediction);
        } catch (Exception e) {
            log.error("Failed to generate prediction: {}", e.getMessage());
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @PostMapping("/predict/batch")
    public ResponseEntity<List<MarketPrediction>> batchGeneratePredictions(
            @RequestBody BatchPredictionRequest request) {

        log.info("REST: Batch generate predictions for {} articles", request.getArticleIds().size());

        try {
            List<MarketPrediction> predictions = marketImpactService.batchGeneratePredictions(
                    request.getArticleIds(),
                    request.getSymbol());
            return ResponseEntity.ok(predictions);
        } catch (Exception e) {
            log.error("Failed to batch generate predictions: {}", e.getMessage());
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class BatchPredictionRequest {
        private List<UUID> articleIds;
        private String symbol;
    }
}