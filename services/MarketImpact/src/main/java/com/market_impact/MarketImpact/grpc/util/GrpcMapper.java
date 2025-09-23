package com.market_impact.MarketImpact.grpc.util;

import com.market_impact.MarketImpact.dto.MarketPredictionDto;
import com.market_impact.MarketImpact.dto.RiskMetricsDto;

// Generated gRPC classes - these will be available after mvn compile
import com.market_impact.grpc.*;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.List;

@Component
public class GrpcMapper {

    // ==================== Common Type Conversions ====================

    public java.util.UUID toUUID(com.market_impact.grpc.UUID grpcUuid) {
        return java.util.UUID.fromString(grpcUuid.getValue());
    }

    public com.market_impact.grpc.UUID toGrpcUUID(UUID uuid) {
        return com.market_impact.grpc.UUID.newBuilder()
                .setValue(uuid.toString())
                .build();
    }

    public BigDecimal toDecimal(Decimal grpcDecimal) {
        return new BigDecimal(grpcDecimal.getValue());
    }

    public Decimal toGrpcDecimal(BigDecimal decimal) {
        return Decimal.newBuilder()
                .setValue(decimal.toString())
                .build();
    }

    public LocalDateTime toLocalDateTime(com.google.protobuf.Timestamp timestamp) {
        return LocalDateTime.ofInstant(
                Instant.ofEpochSecond(timestamp.getSeconds(), timestamp.getNanos()),
                ZoneOffset.UTC
        );
    }

    public com.google.protobuf.Timestamp toGrpcTimestamp(LocalDateTime dateTime) {
        if (dateTime == null) return null;
        Instant instant = dateTime.toInstant(ZoneOffset.UTC);
        return com.google.protobuf.Timestamp.newBuilder()
                .setSeconds(instant.getEpochSecond())
                .setNanos(instant.getNano())
                .build();
    }

    // ==================== Market Prediction Conversions ====================

    public MarketPredictionDto.Request toCreatePredictionDto(CreateMarketPredictionRequest request) {
        return MarketPredictionDto.Request.builder()
                .articleId(toUUID(request.getArticleId()))
                .symbol(request.getSymbol())
                .predictedChangePercent(toDecimal(request.getPredictedChangePercent()))
                .direction(request.getDirection().name())
                .confidence(toDecimal(request.getConfidence()))
                .impactScore(toDecimal(request.getImpactScore()))
                .modelType(request.getModelType())
                .predictionTimestamp(request.hasPredictionTimestamp() ?
                        toLocalDateTime(request.getPredictionTimestamp()) : null)
                .build();
    }

    public MarketPredictionDto.Request toUpdatePredictionDto(UpdateMarketPredictionRequest request) {
        MarketPredictionDto.Request.MarketPredictionRequestBuilder builder =
                MarketPredictionDto.Request.builder();

        if (request.hasArticleId()) {
            builder.articleId(toUUID(request.getArticleId()));
        }
        if (!request.getSymbol().isEmpty()) {
            builder.symbol(request.getSymbol());
        }
        if (request.hasPredictedChangePercent()) {
            builder.predictedChangePercent(toDecimal(request.getPredictedChangePercent()));
        }
        if (request.getDirection() != Direction.DIRECTION_UNSPECIFIED) {
            builder.direction(request.getDirection().name());
        }
        if (request.hasConfidence()) {
            builder.confidence(toDecimal(request.getConfidence()));
        }
        if (request.hasImpactScore()) {
            builder.impactScore(toDecimal(request.getImpactScore()));
        }
        if (!request.getModelType().isEmpty()) {
            builder.modelType(request.getModelType());
        }
        if (request.hasPredictionTimestamp()) {
            builder.predictionTimestamp(toLocalDateTime(request.getPredictionTimestamp()));
        }

        return builder.build();
    }

    public MarketPrediction toGrpcPrediction(MarketPredictionDto.Response dto) {
        MarketPrediction.Builder builder = MarketPrediction.newBuilder()
                .setId(toGrpcUUID(dto.getId()))
                .setArticleId(toGrpcUUID(dto.getArticleId()))
                .setSymbol(dto.getSymbol())
                .setPredictedChangePercent(toGrpcDecimal(dto.getPredictedChangePercent()))
                .setDirection(Direction.valueOf(dto.getDirection()))
                .setConfidence(toGrpcDecimal(dto.getConfidence()))
                .setImpactScore(toGrpcDecimal(dto.getImpactScore()))
                .setModelType(dto.getModelType());

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
                    .map(this::toGrpcRiskMetrics)
                    .toList();
            builder.addAllRiskMetrics(grpcRiskMetrics);
        }

        return builder.build();
    }

    public MarketPredictionSummary toGrpcPredictionSummary(MarketPredictionDto.Summary dto) {
        MarketPredictionSummary.Builder builder = MarketPredictionSummary.newBuilder()
                .setId(toGrpcUUID(dto.getId()))
                .setSymbol(dto.getSymbol())
                .setPredictedChangePercent(toGrpcDecimal(dto.getPredictedChangePercent()))
                .setDirection(Direction.valueOf(dto.getDirection()))
                .setConfidence(toGrpcDecimal(dto.getConfidence()))
                .setModelType(dto.getModelType());

        if (dto.getPredictionTimestamp() != null) {
            builder.setPredictionTimestamp(toGrpcTimestamp(dto.getPredictionTimestamp()));
        }

        return builder.build();
    }

    // ==================== Risk Metrics Conversions ====================

    public RiskMetrics toGrpcRiskMetrics(RiskMetricsDto.Response dto) {
        RiskMetrics.Builder builder = RiskMetrics.newBuilder()
                .setId(toGrpcUUID(dto.getId()))
                .setPredictionId(toGrpcUUID(dto.getPredictionId()))
                .setSymbol(dto.getSymbol())
                .setVar951day(toGrpcDecimal(dto.getVar951day()))
                .setHistoricalVolatility30d(toGrpcDecimal(dto.getHistoricalVolatility30d()))
                .setMarketCorrelation(toGrpcDecimal(dto.getMarketCorrelation()))
                .setRiskLevel(RiskLevel.valueOf(dto.getRiskLevel()));

        if (dto.getCreatedAt() != null) {
            builder.setCreatedAt(toGrpcTimestamp(dto.getCreatedAt()));
        }
        if (dto.getUpdatedAt() != null) {
            builder.setUpdatedAt(toGrpcTimestamp(dto.getUpdatedAt()));
        }

        return builder.build();
    }

    // ==================== Pagination Conversions ====================

    public PageRequest toPageRequest(com.market_impact.grpc.PageRequest grpcPageRequest) {
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

    public com.market_impact.grpc.PageResponse toGrpcPageResponse(Page<?> page) {
        return com.market_impact.grpc.PageResponse.newBuilder()
                .setPage(page.getNumber())
                .setSize(page.getSize())
                .setTotalPages(page.getTotalPages())
                .setTotalElements(page.getTotalElements())
                .setFirst(page.isFirst())
                .setLast(page.isLast())
                .build();
    }

    // ==================== Date Conversions ====================

    public java.time.LocalDate toLocalDate(Date grpcDate) {
        return java.time.LocalDate.of(grpcDate.getYear(), grpcDate.getMonth(), grpcDate.getDay());
    }

    public Date toGrpcDate(java.time.LocalDate date) {
        return Date.newBuilder()
                .setYear(date.getYear())
                .setMonth(date.getMonthValue())
                .setDay(date.getDayOfMonth())
                .build();
    }

    // ==================== Helper Methods ====================

    public String directionToString(Direction direction) {
        return direction == Direction.DIRECTION_UNSPECIFIED ? null : direction.name();
    }

    public Direction stringToDirection(String direction) {
        if (direction == null || direction.isEmpty()) {
            return Direction.DIRECTION_UNSPECIFIED;
        }
        return Direction.valueOf(direction);
    }

    public String riskLevelToString(RiskLevel riskLevel) {
        return riskLevel == RiskLevel.RISK_LEVEL_UNSPECIFIED ? null : riskLevel.name();
    }

    public RiskLevel stringToRiskLevel(String riskLevel) {
        if (riskLevel == null || riskLevel.isEmpty()) {
            return RiskLevel.RISK_LEVEL_UNSPECIFIED;
        }
        return RiskLevel.valueOf(riskLevel);
    }

    // ==================== Validation Helpers ====================

    public boolean isValidUUID(String uuid) {
        try {
            UUID.fromString(uuid);
            return true;
        } catch (IllegalArgumentException e) {
            return false;
        }
    }

    public boolean isValidDecimal(String decimal) {
        try {
            new BigDecimal(decimal);
            return true;
        } catch (NumberFormatException e) {
            return false;
        }
    }

    // ==================== Error Handling ====================

    public RuntimeException createValidationError(String message) {
        return new IllegalArgumentException("Validation error: " + message);
    }

    public RuntimeException createConversionError(String message, Throwable cause) {
        return new RuntimeException("Conversion error: " + message, cause);
    }
}