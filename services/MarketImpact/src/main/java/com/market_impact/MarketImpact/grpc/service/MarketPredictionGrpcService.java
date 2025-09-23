package com.market_impact.MarketImpact.grpc.service;

import com.market_impact.MarketImpact.Services.MarketPredictionService;
import com.market_impact.MarketImpact.dto.MarketPredictionDto;
import com.market_impact.grpc.*;
import com.market_impact.MarketImpact.grpc.util.GrpcMapper;
import com.google.protobuf.Empty;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.grpc.server.service.GrpcService;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@GrpcService
@RequiredArgsConstructor
@Slf4j
public class MarketPredictionGrpcService extends MarketPredictionServiceGrpc.MarketPredictionServiceImplBase {

    private final MarketPredictionService marketPredictionService;
    private final GrpcMapper grpcMapper;

    @Override
    public void createPrediction(CreateMarketPredictionRequest request,
                                 StreamObserver<MarketPrediction> responseObserver) {
        try {
            log.info("Creating prediction for symbol: {}", request.getSymbol());

            MarketPredictionDto.Request dto = grpcMapper.toCreatePredictionDto(request);
            MarketPredictionDto.Response response = marketPredictionService.createPrediction(dto);
            MarketPrediction grpcResponse = grpcMapper.toGrpcPrediction(response);

            responseObserver.onNext(grpcResponse);
            responseObserver.onCompleted();

            log.info("Successfully created prediction with ID: {}", response.getId());
        } catch (Exception e) {
            log.error("Error creating prediction: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    @Override
    public void getPrediction(com.market_impact.grpc.GetMarketPredictionRequest request,
                              StreamObserver<GetMarketPredictionResponse> responseObserver) {
        try {
            log.debug("Getting prediction by ID: {}", request.getId().getValue());

            UUID predictionId = UUID.fromString(request.getId().getValue());
            Optional<MarketPredictionDto.Response> prediction =
                    marketPredictionService.getPredictionById(predictionId);

            if (prediction.isPresent()) {
                GetMarketPredictionResponse response = GetMarketPredictionResponse.newBuilder()
                        .setPrediction(grpcMapper.toGrpcPrediction(prediction.get()))
                        .build();
                responseObserver.onNext(response);
            } else {
                responseObserver.onError(new RuntimeException("Prediction not found"));
            }
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting prediction: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    @Override
    public void updatePrediction(UpdateMarketPredictionRequest request,
                                 StreamObserver<MarketPrediction> responseObserver) {
        try {
            log.info("Updating prediction with ID: {}", request.getId().getValue());

            UUID predictionId = UUID.fromString(request.getId().getValue());
            MarketPredictionDto.Request dto = grpcMapper.toUpdatePredictionDto(request);
            MarketPredictionDto.Response response =
                    marketPredictionService.updatePrediction(predictionId, dto);
            MarketPrediction grpcResponse = grpcMapper.toGrpcPrediction(response);

            responseObserver.onNext(grpcResponse);
            responseObserver.onCompleted();

            log.info("Successfully updated prediction with ID: {}", predictionId);
        } catch (Exception e) {
            log.error("Error updating prediction: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    @Override
    public void deletePrediction(DeleteMarketPredictionRequest request,
                                 StreamObserver<Empty> responseObserver) {
        try {
            log.info("Deleting prediction with ID: {}", request.getId().getValue());

            UUID predictionId = UUID.fromString(request.getId().getValue());
            marketPredictionService.deletePrediction(predictionId);

            responseObserver.onNext(Empty.getDefaultInstance());
            responseObserver.onCompleted();

            log.info("Successfully deleted prediction with ID: {}", predictionId);
        } catch (Exception e) {
            log.error("Error deleting prediction: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    @Override
    public void getPredictionsBySymbol(GetPredictionsBySymbolRequest request,
                                       StreamObserver<GetPredictionsBySymbolResponse> responseObserver) {
        try {
            log.debug("Getting predictions for symbol: {}", request.getSymbol());

            if (request.hasPageRequest()) {
                // Handle paginated request
                PageRequest pageRequest = grpcMapper.toPageRequest(request.getPageRequest());
                Page<MarketPredictionDto.Response> pagedPredictions =
                        marketPredictionService.getPredictionsBySymbolPaged(request.getSymbol(), pageRequest);

                List<MarketPrediction> grpcPredictions = pagedPredictions.getContent()
                        .stream()
                        .map(grpcMapper::toGrpcPrediction)
                        .toList();

                PageResponse pageResponse = grpcMapper.toGrpcPageResponse(pagedPredictions);

                GetPredictionsBySymbolResponse response = GetPredictionsBySymbolResponse.newBuilder()
                        .addAllPredictions(grpcPredictions)
                        .setPageResponse(pageResponse)
                        .build();

                responseObserver.onNext(response);
            } else {
                // Handle non-paginated request
                List<MarketPredictionDto.Response> predictions =
                        marketPredictionService.getPredictionsBySymbol(request.getSymbol());

                List<MarketPrediction> grpcPredictions = predictions.stream()
                        .map(grpcMapper::toGrpcPrediction)
                        .toList();

                GetPredictionsBySymbolResponse response = GetPredictionsBySymbolResponse.newBuilder()
                        .addAllPredictions(grpcPredictions)
                        .build();

                responseObserver.onNext(response);
            }

            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting predictions by symbol: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    @Override
    public void getPredictionsSummary(GetPredictionsSummaryRequest request,
                                      StreamObserver<GetPredictionsSummaryResponse> responseObserver) {
        try {
            log.debug("Getting predictions summary for symbol: {}", request.getSymbol());

            List<MarketPredictionDto.Summary> summaries =
                    marketPredictionService.getPredictionsSummaryBySymbol(request.getSymbol());

            List<MarketPredictionSummary> grpcSummaries = summaries.stream()
                    .map(grpcMapper::toGrpcPredictionSummary)
                    .toList();

            GetPredictionsSummaryResponse response = GetPredictionsSummaryResponse.newBuilder()
                    .addAllSummaries(grpcSummaries)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting predictions summary: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    @Override
    public void getRecentPredictions(GetRecentPredictionsRequest request,
                                     StreamObserver<GetRecentPredictionsResponse> responseObserver) {
        try {
            log.debug("Getting recent predictions for {} hours", request.getHours());

            List<MarketPredictionDto.Response> predictions =
                    marketPredictionService.getRecentPredictions(request.getHours());

            List<MarketPrediction> grpcPredictions = predictions.stream()
                    .map(grpcMapper::toGrpcPrediction)
                    .toList();

            GetRecentPredictionsResponse response = GetRecentPredictionsResponse.newBuilder()
                    .addAllPredictions(grpcPredictions)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting recent predictions: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    @Override
    public void getHighConfidencePredictions(GetHighConfidencePredictionsRequest request,
                                             StreamObserver<GetHighConfidencePredictionsResponse> responseObserver) {
        try {
            double minConfidence = Double.parseDouble(request.getMinConfidence().getValue());
            log.debug("Getting high confidence predictions with min confidence: {}", minConfidence);

            List<MarketPredictionDto.Response> predictions =
                    marketPredictionService.getHighConfidencePredictions(minConfidence);

            List<MarketPrediction> grpcPredictions = predictions.stream()
                    .map(grpcMapper::toGrpcPrediction)
                    .toList();

            GetHighConfidencePredictionsResponse response = GetHighConfidencePredictionsResponse.newBuilder()
                    .addAllPredictions(grpcPredictions)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting high confidence predictions: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    @Override
    public void getPredictionsByTimeRange(GetPredictionsByTimeRangeRequest request,
                                          StreamObserver<GetPredictionsByTimeRangeResponse> responseObserver) {
        try {
            log.debug("Getting predictions by time range for symbol: {}", request.getSymbol());

            java.time.LocalDateTime startTime = grpcMapper.toLocalDateTime(request.getTimeRange().getStartTime());
            java.time.LocalDateTime endTime = grpcMapper.toLocalDateTime(request.getTimeRange().getEndTime());

            List<MarketPredictionDto.Response> predictions =
                    marketPredictionService.getPredictionsByTimeRange(request.getSymbol(), startTime, endTime);

            List<MarketPrediction> grpcPredictions = predictions.stream()
                    .map(grpcMapper::toGrpcPrediction)
                    .toList();

            GetPredictionsByTimeRangeResponse response = GetPredictionsByTimeRangeResponse.newBuilder()
                    .addAllPredictions(grpcPredictions)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting predictions by time range: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    @Override
    public void getAvailableSymbols(GetAvailableSymbolsRequest request,
                                    StreamObserver<GetAvailableSymbolsResponse> responseObserver) {
        try {
            log.debug("Getting available symbols");

            List<String> symbols = marketPredictionService.getAvailableSymbols();

            GetAvailableSymbolsResponse response = GetAvailableSymbolsResponse.newBuilder()
                    .addAllSymbols(symbols)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting available symbols: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    @Override
    public void getLatestPrediction(GetLatestPredictionRequest request,
                                    StreamObserver<GetLatestPredictionResponse> responseObserver) {
        try {
            log.debug("Getting latest prediction for symbol: {}", request.getSymbol());

            Optional<MarketPredictionDto.Response> prediction =
                    marketPredictionService.getLatestPredictionForSymbol(request.getSymbol());

            if (prediction.isPresent()) {
                GetLatestPredictionResponse response = GetLatestPredictionResponse.newBuilder()
                        .setPrediction(grpcMapper.toGrpcPrediction(prediction.get()))
                        .build();
                responseObserver.onNext(response);
            } else {
                responseObserver.onError(new RuntimeException("No prediction found for symbol: " + request.getSymbol()));
            }
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting latest prediction: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }
}