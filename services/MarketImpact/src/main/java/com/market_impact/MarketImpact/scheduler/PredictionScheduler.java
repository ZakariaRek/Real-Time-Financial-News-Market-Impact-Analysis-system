package com.market_impact.MarketImpact.scheduler;

import com.market_impact.MarketImpact.Services.MarketImpactPredictionService;
import com.market_impact.MarketImpact.Repositories.MarketPredictionRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

@Component
@RequiredArgsConstructor
@Slf4j
public class PredictionScheduler {

    private final MarketImpactPredictionService marketImpactService;

    /**
     * Run every 5 minutes to generate predictions for new sentiment analyses
     */
    @Scheduled(fixedDelay = 300000) // 5 minutes
    public void generatePredictionsForNewSentiment() {
        log.info("Running scheduled prediction generation");

        try {
            // Logic to fetch articles with sentiment but no predictions
            // Then generate predictions

            log.info("Scheduled prediction generation completed");
        } catch (Exception e) {
            log.error("Scheduled prediction generation failed: {}", e.getMessage());
        }
    }
}