// services/MarketImpact/src/main/java/com/market_impact/MarketImpact/controller/SP500MarketImpactController.java
package com.market_impact.MarketImpact.controller;

import com.market_impact.MarketImpact.Services.SP500MarketImpactService;
import com.market_impact.MarketImpact.dto.MarketImpactDto;
import com.market_impact.MarketImpact.dto.MarketSentimentSummary;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/sp500/market-impact")
@RequiredArgsConstructor
@Slf4j
@CrossOrigin(origins = "*") // Configure appropriately for production
public class SP500MarketImpactController {

    private final SP500MarketImpactService sp500Service;

    /**
     * Get current market impact for all S&P 500 stocks
     */
    @GetMapping
    public ResponseEntity<List<MarketImpactDto>> getSP500MarketImpact() {
        log.info("REST: Get S&P 500 market impact");

        List<MarketImpactDto> impacts = sp500Service.getSP500MarketImpact();
        return ResponseEntity.ok(impacts);
    }

    /**
     * Get market impact for specific symbols
     */
    @PostMapping("/symbols")
    public ResponseEntity<List<MarketImpactDto>> getMarketImpactForSymbols(
            @RequestBody List<String> symbols) {
        log.info("REST: Get market impact for {} symbols", symbols.size());

        List<MarketImpactDto> impacts = sp500Service.getMarketImpactForSymbols(symbols);
        return ResponseEntity.ok(impacts);
    }

    /**
     * Get top movers (highest impact scores)
     */
    @GetMapping("/top-movers")
    public ResponseEntity<List<MarketImpactDto>> getTopMovers(
            @RequestParam(defaultValue = "20") int limit) {
        log.info("REST: Get top {} movers", limit);

        List<MarketImpactDto> topMovers = sp500Service.getTopMovers(limit);
        return ResponseEntity.ok(topMovers);
    }

    /**
     * Get market impact grouped by direction
     */
    @GetMapping("/by-direction")
    public ResponseEntity<Map<String, List<MarketImpactDto>>> getMarketImpactByDirection() {
        log.info("REST: Get market impact by direction");

        Map<String, List<MarketImpactDto>> impactsByDirection =
                sp500Service.getMarketImpactByDirection();
        return ResponseEntity.ok(impactsByDirection);
    }

    /**
     * Get market sentiment summary
     */
    @GetMapping("/summary")
    public ResponseEntity<MarketSentimentSummary> getMarketSentimentSummary() {
        log.info("REST: Get market sentiment summary");

        MarketSentimentSummary summary = sp500Service.getMarketSentimentSummary();
        return ResponseEntity.ok(summary);
    }

    /**
     * Server-Sent Events endpoint for real-time market impact updates
     */
    @GetMapping(value = "/stream", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public SseEmitter streamMarketImpact(
            @RequestParam(defaultValue = "3600000") Long timeout) { // 1 hour default

        log.info("REST: SSE connection requested with timeout: {}ms", timeout);

        SseEmitter emitter = sp500Service.registerEmitter(timeout);

        // Send initial data
        try {
            List<MarketImpactDto> currentImpacts = sp500Service.getSP500MarketImpact();
            emitter.send(SseEmitter.event()
                    .name("initial-data")
                    .data(currentImpacts));

            MarketSentimentSummary summary = sp500Service.getMarketSentimentSummary();
            emitter.send(SseEmitter.event()
                    .name("market-summary")
                    .data(summary));

        } catch (Exception e) {
            log.error("Failed to send initial SSE data: {}", e.getMessage());
            emitter.completeWithError(e);
        }

        return emitter;
    }
}