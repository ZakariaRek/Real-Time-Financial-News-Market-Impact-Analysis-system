package com.market_impact.MarketImpact.Services;

import com.market_impact.MarketImpact.Mappers.MarketPredictionMapper;
import com.market_impact.MarketImpact.dto.MarketPredictionDto;
import com.market_impact.MarketImpact.entity.MarketPrediction;
import com.market_impact.MarketImpact.Repositories.MarketPredictionRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import lombok.extern.slf4j.XSlf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
@Transactional
public class MarketPredictionService {

    private final MarketPredictionRepository marketPredictionRepository;
    private final MarketPredictionMapper marketPredictionMapper;

    public MarketPredictionDto.Response createPrediction(MarketPredictionDto.Request request) {
//        // log.info("Creating market prediction for symbol: {}", request.getSymbol());

        MarketPrediction prediction = marketPredictionMapper.toEntity(request);
        MarketPrediction savedPrediction = marketPredictionRepository.save(prediction);

        // // log.info("Created market prediction with ID: {}", savedPrediction.getId());
        return marketPredictionMapper.toResponse(savedPrediction);
    }

    @Transactional(readOnly = true)
    public Optional<MarketPredictionDto.Response> getPredictionById(UUID id) {
        // log.debug("Fetching market prediction by ID: {}", id);

        return marketPredictionRepository.findById(id)
                .map(marketPredictionMapper::toResponse);
    }

    @Transactional(readOnly = true)
    public List<MarketPredictionDto.Response> getPredictionsBySymbol(String symbol) {
        // log.debug("Fetching market predictions for symbol: {}", symbol);

        List<MarketPrediction> predictions = marketPredictionRepository.findBySymbol(symbol);
        return predictions.stream()
                .map(marketPredictionMapper::toResponse)
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public List<MarketPredictionDto.Summary> getPredictionsSummaryBySymbol(String symbol) {
        // log.debug("Fetching predictions summary for symbol: {}", symbol);

        List<MarketPrediction> predictions = marketPredictionRepository.findBySymbol(symbol);
        return predictions.stream()
                .map(marketPredictionMapper::toSummary)
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public Page<MarketPredictionDto.Response> getPredictionsBySymbolPaged(String symbol, Pageable pageable) {
        // log.debug("Fetching paginated market predictions for symbol: {}", symbol);

        Page<MarketPrediction> predictions = marketPredictionRepository
                .findBySymbolOrderByTimestampDesc(symbol, pageable);
        return predictions.map(marketPredictionMapper::toResponse);
    }

    @Transactional(readOnly = true)
    public List<MarketPredictionDto.Response> getRecentPredictions(int hours) {
        // log.debug("Fetching recent predictions from last {} hours", hours);

        LocalDateTime cutoffTime = LocalDateTime.now().minusHours(hours);
        List<MarketPrediction> predictions = marketPredictionRepository.findRecentPredictions(cutoffTime);

        return predictions.stream()
                .map(marketPredictionMapper::toResponse)
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public List<MarketPredictionDto.Response> getHighConfidencePredictions(double minConfidence) {
        // log.debug("Fetching high confidence predictions with confidence >= {}", minConfidence);

        List<MarketPrediction> predictions = marketPredictionRepository
                .findHighConfidencePredictions(minConfidence);

        return predictions.stream()
                .map(marketPredictionMapper::toResponse)
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public List<MarketPredictionDto.Response> getPredictionsByTimeRange(
            String symbol, LocalDateTime startTime, LocalDateTime endTime) {
        // log.debug("Fetching predictions for symbol {} between {} and {}", symbol, startTime, endTime);

        List<MarketPrediction> predictions = marketPredictionRepository
                .findBySymbolAndTimestampRange(symbol, startTime, endTime);

        return predictions.stream()
                .map(marketPredictionMapper::toResponse)
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public List<String> getAvailableSymbols() {
        // log.debug("Fetching available symbols");
        return marketPredictionRepository.findDistinctSymbols();
    }

    @Transactional(readOnly = true)
    public Optional<MarketPredictionDto.Response> getLatestPredictionForSymbol(String symbol) {
        // log.debug("Fetching latest prediction for symbol: {}", symbol);

        return marketPredictionRepository.findTopBySymbolOrderByPredictionTimestampDesc(symbol)
                .map(marketPredictionMapper::toResponse);
    }

    public MarketPredictionDto.Response updatePrediction(UUID id, MarketPredictionDto.Request request) {
        // log.info("Updating market prediction with ID: {}", id);

        MarketPrediction prediction = marketPredictionRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("Market prediction not found with ID: " + id));

        marketPredictionMapper.updateEntityFromRequest(request, prediction);
        MarketPrediction updatedPrediction = marketPredictionRepository.save(prediction);

        // log.info("Updated market prediction with ID: {}", updatedPrediction.getId());
        return marketPredictionMapper.toResponse(updatedPrediction);
    }

    public void deletePrediction(UUID id) {
        // log.info("Deleting market prediction with ID: {}", id);

        if (!marketPredictionRepository.existsById(id)) {
            throw new RuntimeException("Market prediction not found with ID: " + id);
        }

        marketPredictionRepository.deleteById(id);
        // log.info("Deleted market prediction with ID: {}", id);
    }
}