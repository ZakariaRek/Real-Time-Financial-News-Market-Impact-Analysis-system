package com.market_impact.MarketImpact.grpc.service;

import com.market_impact.MarketImpact.Services.MarketPredictionService;
import com.market_impact.MarketImpact.dto.MarketPredictionDto;
import com.market_impact.MarketImpact.dto.RiskMetricsDto;
import com.google.protobuf.Empty;
import com.google.protobuf.Timestamp;
// REMOVED: import com.market_impact.MarketImpact.entity.MarketPrediction;
import com.market_impact.grpc.*;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Sort;
import org.springframework.grpc.server.service.GrpcService;

import java.math.BigDecimal;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

@GrpcService
@RequiredArgsConstructor
@Slf4j
public class MarketPredictionGrpcService extends MarketPredictionServiceGrpc.MarketPredictionServiceImplBase {

    private final MarketPredictionService marketPredictionService;

    @Override
    public void createPrediction(CreateMarketPredictionRequest request,
                                 StreamObserver<com.market_impact.grpc.MarketPrediction> responseObserver) {
        try {
            MarketPredictionDto.Request dto = MarketPredictionDto.Request.builder()
                    .articleId(UUID.fromString(request.getArticleId().getValue()))
                    .symbol(request.getSymbol())
                    .predictedChangePercent(new BigDecimal(request.getPredictedChangePercent().getValue()))
                    .direction(request.getDirection().name())
                    .confidence(new BigDecimal(request.getConfidence().getValue()))
                    .impactScore(new BigDecimal(request.getImpactScore().getValue()))
                    .modelType(request.getModelType())
                    .predictionTimestamp(request.hasPredictionTimestamp() ?
                            toLocalDateTime(request.getPredictionTimestamp()) : null)
                    .build();

            MarketPredictionDto.Response response = marketPredictionService.createPrediction(dto);
            com.market_impact.grpc.MarketPrediction grpcResponse = mapToMarketPredictionResponse(response);

            responseObserver.onNext(grpcResponse);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error creating prediction: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getPrediction(GetMarketPredictionRequest request,
                              StreamObserver<GetMarketPredictionResponse> responseObserver) {
        try {
            UUID predictionId = UUID.fromString(request.getId().getValue());
            Optional<MarketPredictionDto.Response> prediction =
                    marketPredictionService.getPredictionById(predictionId);

            if (prediction.isPresent()) {
                GetMarketPredictionResponse response = GetMarketPredictionResponse.newBuilder()
                        .setPrediction(mapToMarketPredictionResponse(prediction.get()))
                        .build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("Prediction not found")
                        .asRuntimeException());
            }
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting prediction: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void updatePrediction(UpdateMarketPredictionRequest request,
                                 StreamObserver<com.market_impact.grpc.MarketPrediction> responseObserver) {
        try {
            UUID predictionId = UUID.fromString(request.getId().getValue());

            MarketPredictionDto.Request.RequestBuilder builder =
                    MarketPredictionDto.Request.builder();

            if (request.hasArticleId()) {
                builder.articleId(UUID.fromString(request.getArticleId().getValue()));
            }
            if (!request.getSymbol().isEmpty()) {
                builder.symbol(request.getSymbol());
            }
            if (request.hasPredictedChangePercent()) {
                builder.predictedChangePercent(new BigDecimal(request.getPredictedChangePercent().getValue()));
            }
            if (request.getDirection() != Direction.DIRECTION_UNSPECIFIED) {
                builder.direction(request.getDirection().name());
            }
            if (request.hasConfidence()) {
                builder.confidence(new BigDecimal(request.getConfidence().getValue()));
            }
            if (request.hasImpactScore()) {
                builder.impactScore(new BigDecimal(request.getImpactScore().getValue()));
            }
            if (!request.getModelType().isEmpty()) {
                builder.modelType(request.getModelType());
            }
            if (request.hasPredictionTimestamp()) {
                builder.predictionTimestamp(toLocalDateTime(request.getPredictionTimestamp()));
            }

            MarketPredictionDto.Request dto = builder.build();
            MarketPredictionDto.Response response =
                    marketPredictionService.updatePrediction(predictionId, dto);
            com.market_impact.grpc.MarketPrediction grpcResponse = mapToMarketPredictionResponse(response);

            responseObserver.onNext(grpcResponse);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error updating prediction: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void deletePrediction(DeleteMarketPredictionRequest request,
                                 StreamObserver<Empty> responseObserver) {
        try {
            UUID predictionId = UUID.fromString(request.getId().getValue());
            marketPredictionService.deletePrediction(predictionId);

            responseObserver.onNext(Empty.getDefaultInstance());
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error deleting prediction: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getPredictionsBySymbol(GetPredictionsBySymbolRequest request,
                                       StreamObserver<GetPredictionsBySymbolResponse> responseObserver) {
        try {
            if (request.hasPageRequest()) {
                // Handle paginated request
                PageRequest pageRequest = createPageRequest(request.getPageRequest());
                Page<MarketPredictionDto.Response> pagedPredictions =
                        marketPredictionService.getPredictionsBySymbolPaged(request.getSymbol(), pageRequest);

                List<com.market_impact.grpc.MarketPrediction> grpcPredictions = pagedPredictions.getContent()
                        .stream()
                        .map(this::mapToMarketPredictionResponse)
                        .collect(Collectors.toList());

                com.market_impact.grpc.PageResponse pageResponse = createPageResponse(pagedPredictions);

                GetPredictionsBySymbolResponse response = GetPredictionsBySymbolResponse.newBuilder()
                        .addAllPredictions(grpcPredictions)
                        .setPageResponse(pageResponse)
                        .build();

                responseObserver.onNext(response);
            } else {
                // Handle non-paginated request
                List<MarketPredictionDto.Response> predictions =
                        marketPredictionService.getPredictionsBySymbol(request.getSymbol());

                List<com.market_impact.grpc.MarketPrediction> grpcPredictions = predictions.stream()
                        .map(this::mapToMarketPredictionResponse)
                        .collect(Collectors.toList());

                GetPredictionsBySymbolResponse response = GetPredictionsBySymbolResponse.newBuilder()
                        .addAllPredictions(grpcPredictions)
                        .build();

                responseObserver.onNext(response);
            }

            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting predictions by symbol: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getPredictionsSummary(GetPredictionsSummaryRequest request,
                                      StreamObserver<GetPredictionsSummaryResponse> responseObserver) {
        try {
            List<MarketPredictionDto.Summary> summaries =
                    marketPredictionService.getPredictionsSummaryBySymbol(request.getSymbol());

            List<MarketPredictionSummary> grpcSummaries = summaries.stream()
                    .map(this::mapToMarketPredictionSummaryResponse)
                    .collect(Collectors.toList());

            GetPredictionsSummaryResponse response = GetPredictionsSummaryResponse.newBuilder()
                    .addAllSummaries(grpcSummaries)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting predictions summary: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getRecentPredictions(GetRecentPredictionsRequest request,
                                     StreamObserver<GetRecentPredictionsResponse> responseObserver) {
        try {
            List<MarketPredictionDto.Response> predictions =
                    marketPredictionService.getRecentPredictions(request.getHours());

            List<com.market_impact.grpc.MarketPrediction> grpcPredictions = predictions.stream()
                    .map(this::mapToMarketPredictionResponse)
                    .collect(Collectors.toList());

            GetRecentPredictionsResponse response = GetRecentPredictionsResponse.newBuilder()
                    .addAllPredictions(grpcPredictions)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting recent predictions: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getHighConfidencePredictions(GetHighConfidencePredictionsRequest request,
                                             StreamObserver<GetHighConfidencePredictionsResponse> responseObserver) {
        try {
            double minConfidence = Double.parseDouble(request.getMinConfidence().getValue());

            List<MarketPredictionDto.Response> predictions =
                    marketPredictionService.getHighConfidencePredictions(minConfidence);

            List<com.market_impact.grpc.MarketPrediction> grpcPredictions = predictions.stream()
                    .map(this::mapToMarketPredictionResponse)
                    .collect(Collectors.toList());

            GetHighConfidencePredictionsResponse response = GetHighConfidencePredictionsResponse.newBuilder()
                    .addAllPredictions(grpcPredictions)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting high confidence predictions: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getPredictionsByTimeRange(GetPredictionsByTimeRangeRequest request,
                                          StreamObserver<GetPredictionsByTimeRangeResponse> responseObserver) {
        try {
            LocalDateTime startTime = toLocalDateTime(request.getTimeRange().getStartTime());
            LocalDateTime endTime = toLocalDateTime(request.getTimeRange().getEndTime());

            List<MarketPredictionDto.Response> predictions =
                    marketPredictionService.getPredictionsByTimeRange(request.getSymbol(), startTime, endTime);

            List<com.market_impact.grpc.MarketPrediction> grpcPredictions = predictions.stream()
                    .map(this::mapToMarketPredictionResponse)
                    .collect(Collectors.toList());

            GetPredictionsByTimeRangeResponse response = GetPredictionsByTimeRangeResponse.newBuilder()
                    .addAllPredictions(grpcPredictions)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting predictions by time range: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getAvailableSymbols(GetAvailableSymbolsRequest request,
                                    StreamObserver<GetAvailableSymbolsResponse> responseObserver) {
        try {
            List<String> symbols = marketPredictionService.getAvailableSymbols();

            GetAvailableSymbolsResponse response = GetAvailableSymbolsResponse.newBuilder()
                    .addAllSymbols(symbols)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting available symbols: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getLatestPrediction(GetLatestPredictionRequest request,
                                    StreamObserver<GetLatestPredictionResponse> responseObserver) {
        try {
            Optional<MarketPredictionDto.Response> prediction =
                    marketPredictionService.getLatestPredictionForSymbol(request.getSymbol());

            if (prediction.isPresent()) {
                GetLatestPredictionResponse response = GetLatestPredictionResponse.newBuilder()
                        .setPrediction(mapToMarketPredictionResponse(prediction.get()))
                        .build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("No prediction found for symbol: " + request.getSymbol())
                        .asRuntimeException());
            }
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting latest prediction: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    // Private helper methods for mapping

    private com.market_impact.grpc.MarketPrediction mapToMarketPredictionResponse(MarketPredictionDto.Response dto) {
        com.market_impact.grpc.MarketPrediction.Builder builder = com.market_impact.grpc.MarketPrediction.newBuilder()
                .setId(createGrpcUUID(dto.getId()))
                .setSymbol(dto.getSymbol())
                .setPredictedChangePercent(createGrpcDecimal(dto.getPredictedChangePercent()))
                .setDirection(Direction.valueOf(dto.getDirection()))
                .setConfidence(createGrpcDecimal(dto.getConfidence()))
                .setImpactScore(createGrpcDecimal(dto.getImpactScore()))
                .setModelType(dto.getModelType());

        // Handle nullable articleId
        if (dto.getArticleId() != null) {
            builder.setArticleId(createGrpcUUID(dto.getArticleId()));
        }

        if (dto.getPredictionTimestamp() != null) {
            builder.setPredictionTimestamp(toGrpcTimestamp(dto.getPredictionTimestamp()));
        }
        if (dto.getCreatedAt() != null) {
            builder.setCreatedAt(toGrpcTimestamp(dto.getCreatedAt()));
        }
        if (dto.getUpdatedAt() != null) {
            builder.setUpdatedAt(toGrpcTimestamp(dto.getUpdatedAt()));
        }
        if (dto.getRiskMetrics() != null) {
            List<RiskMetrics> grpcRiskMetrics = dto.getRiskMetrics().stream()
                    .map(this::mapToRiskMetricsResponse)
                    .collect(Collectors.toList());
            builder.addAllRiskMetrics(grpcRiskMetrics);
        }

        return builder.build();
    }

    private MarketPredictionSummary mapToMarketPredictionSummaryResponse(MarketPredictionDto.Summary dto) {
        MarketPredictionSummary.Builder builder = MarketPredictionSummary.newBuilder()
                .setId(createGrpcUUID(dto.getId()))
                .setSymbol(dto.getSymbol())
                .setPredictedChangePercent(createGrpcDecimal(dto.getPredictedChangePercent()))
                .setDirection(Direction.valueOf(dto.getDirection()))
                .setConfidence(createGrpcDecimal(dto.getConfidence()))
                .setModelType(dto.getModelType());

        if (dto.getPredictionTimestamp() != null) {
            builder.setPredictionTimestamp(toGrpcTimestamp(dto.getPredictionTimestamp()));
        }

        return builder.build();
    }

    private RiskMetrics mapToRiskMetricsResponse(RiskMetricsDto.Response dto) {
        RiskMetrics.Builder builder = RiskMetrics.newBuilder()
                .setId(createGrpcUUID(dto.getId()))
                .setSymbol(dto.getSymbol())
                .setVar951Day(createGrpcDecimal(dto.getVar951day()))
                .setHistoricalVolatility30D(createGrpcDecimal(dto.getHistoricalVolatility30d()))
                .setMarketCorrelation(createGrpcDecimal(dto.getMarketCorrelation()))
                .setRiskLevel(RiskLevel.valueOf(dto.getRiskLevel()));

        // Handle nullable predictionId
        if (dto.getPredictionId() != null) {
            builder.setPredictionId(createGrpcUUID(dto.getPredictionId()));
        }

        if (dto.getCreatedAt() != null) {
            builder.setCreatedAt(toGrpcTimestamp(dto.getCreatedAt()));
        }
        if (dto.getUpdatedAt() != null) {
            builder.setUpdatedAt(toGrpcTimestamp(dto.getUpdatedAt()));
        }

        return builder.build();
    }   
    // Common conversion helper methods

    private com.market_impact.grpc.UUID createGrpcUUID(UUID uuid) {
        return com.market_impact.grpc.UUID.newBuilder()
                .setValue(uuid.toString())
                .build();
    }

    private Decimal createGrpcDecimal(BigDecimal decimal) {
        return Decimal.newBuilder()
                .setValue(decimal.toString())
                .build();
    }

    private LocalDateTime toLocalDateTime(Timestamp timestamp) {
        return LocalDateTime.ofInstant(
                Instant.ofEpochSecond(timestamp.getSeconds(), timestamp.getNanos()),
                ZoneOffset.UTC
        );
    }

    private Timestamp toGrpcTimestamp(LocalDateTime dateTime) {
        if (dateTime == null) return null;
        Instant instant = dateTime.toInstant(ZoneOffset.UTC);
        return Timestamp.newBuilder()
                .setSeconds(instant.getEpochSecond())
                .setNanos(instant.getNano())
                .build();
    }

    private PageRequest createPageRequest(com.market_impact.grpc.PageRequest grpcPageRequest) {
        int page = grpcPageRequest.getPage();
        int size = grpcPageRequest.getSize();
        String sortStr = grpcPageRequest.getSort();

        if (sortStr.isEmpty()) {
            return PageRequest.of(page, size);
        }

        // Parse sort string (e.g., "predictionTimestamp,desc")
        String[] sortParts = sortStr.split(",");
        String property = sortParts[0];
        Sort.Direction direction = sortParts.length > 1 && "desc".equalsIgnoreCase(sortParts[1])
                ? Sort.Direction.DESC : Sort.Direction.ASC;

        Sort sort = Sort.by(direction, property);
        return PageRequest.of(page, size, sort);
    }

    private com.market_impact.grpc.PageResponse createPageResponse(Page<?> page) {
        return com.market_impact.grpc.PageResponse.newBuilder()
                .setPage(page.getNumber())
                .setSize(page.getSize())
                .setTotalPages(page.getTotalPages())
                .setTotalElements(page.getTotalElements())
                .setFirst(page.isFirst())
                .setLast(page.isLast())
                .build();
    }
}