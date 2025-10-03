package com.alert_signals.AlertSignal.grpc;

import com.alert_signals.AlertSignal.entity.TradingSignal;
import com.alert_signals.AlertSignal.service.TradingSignalService;
import com.alert_signals.AlertSignal.grpc.generated.CreateSignalRequest;
import com.alert_signals.AlertSignal.grpc.generated.DeleteSignalRequest;
import com.alert_signals.AlertSignal.grpc.generated.DeleteSignalResponse;
import com.alert_signals.AlertSignal.grpc.generated.GetSignalRequest;
import com.alert_signals.AlertSignal.grpc.generated.GetSignalsByDateRangeRequest;
import com.alert_signals.AlertSignal.grpc.generated.GetSignalsByStatusRequest;
import com.alert_signals.AlertSignal.grpc.generated.GetSignalsBySymbolRequest;
import com.alert_signals.AlertSignal.grpc.generated.SignalListResponse;
import com.alert_signals.AlertSignal.grpc.generated.SignalResponse;
import com.alert_signals.AlertSignal.grpc.generated.UpdateSignalRequest;

import com.google.protobuf.Timestamp;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.grpc.server.service.GrpcService;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

@GrpcService
@RequiredArgsConstructor
@Slf4j
public class TradingSignalGrpcServiceImpl extends com.alert_signals.AlertSignal.grpc.generated.TradingSignalGrpcServiceGrpc.TradingSignalGrpcServiceImplBase {

    private final TradingSignalService tradingSignalService;

    @Override
    public void createSignal(CreateSignalRequest request, StreamObserver<SignalResponse> responseObserver) {
        try {
            TradingSignal signal = TradingSignal.builder()
                    .predictionId(request.getPredictionId().isEmpty() ? null : UUID.fromString(request.getPredictionId()))
                    .symbol(request.getSymbol())
                    .signalType(request.getSignalType())
                    .direction(request.getDirection())
                    .strength(BigDecimal.valueOf(request.getStrength()))
                    .confidence(BigDecimal.valueOf(request.getConfidence()))
                    .status(request.getStatus())
                    .build();

            TradingSignal savedSignal = tradingSignalService.createSignal(signal);
            SignalResponse response = mapToSignalResponse(savedSignal);

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error creating trading signal", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error creating trading signal: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getSignal(GetSignalRequest request, StreamObserver<SignalResponse> responseObserver) {
        try {
            UUID signalId = UUID.fromString(request.getId());
            Optional<TradingSignal> signal = tradingSignalService.getSignalById(signalId);

            if (signal.isPresent()) {
                SignalResponse response = mapToSignalResponse(signal.get());
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("Trading signal not found")
                        .asRuntimeException());
            }
        } catch (Exception e) {
            log.error("Error getting trading signal", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting trading signal: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getSignalsBySymbol(GetSignalsBySymbolRequest request, StreamObserver<SignalListResponse> responseObserver) {
        try {
            List<TradingSignal> signals = tradingSignalService.getSignalsBySymbol(request.getSymbol());
            List<SignalResponse> signalResponses = signals.stream()
                    .map(this::mapToSignalResponse)
                    .collect(Collectors.toList());

            SignalListResponse response = SignalListResponse.newBuilder()
                    .addAllSignals(signalResponses)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting signals by symbol", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting signals by symbol: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getSignalsByStatus(GetSignalsByStatusRequest request, StreamObserver<SignalListResponse> responseObserver) {
        try {
            List<TradingSignal> signals = tradingSignalService.getSignalsByStatus(request.getStatus());
            List<SignalResponse> signalResponses = signals.stream()
                    .map(this::mapToSignalResponse)
                    .collect(Collectors.toList());

            SignalListResponse response = SignalListResponse.newBuilder()
                    .addAllSignals(signalResponses)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting signals by status", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting signals by status: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void updateSignal(UpdateSignalRequest request, StreamObserver<SignalResponse> responseObserver) {
        try {
            UUID signalId = UUID.fromString(request.getId());
            Optional<TradingSignal> existingSignal = tradingSignalService.getSignalById(signalId);

            if (existingSignal.isEmpty()) {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("Trading signal not found")
                        .asRuntimeException());
                return;
            }

            TradingSignal signal = TradingSignal.builder()
                    .id(signalId)
                    .predictionId(request.getPredictionId().isEmpty() ? null : UUID.fromString(request.getPredictionId()))
                    .symbol(request.getSymbol())
                    .signalType(request.getSignalType())
                    .direction(request.getDirection())
                    .strength(BigDecimal.valueOf(request.getStrength()))
                    .confidence(BigDecimal.valueOf(request.getConfidence()))
                    .status(request.getStatus())
                    .actualReturnPercent(BigDecimal.valueOf(request.getActualReturnPercent()))
                    .generatedAt(existingSignal.get().getGeneratedAt()) // Keep original timestamp
                    .build();

            TradingSignal updatedSignal = tradingSignalService.updateSignal(signal);
            SignalResponse response = mapToSignalResponse(updatedSignal);

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error updating trading signal", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error updating trading signal: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void deleteSignal(DeleteSignalRequest request, StreamObserver<DeleteSignalResponse> responseObserver) {
        try {
            UUID signalId = UUID.fromString(request.getId());
            Optional<TradingSignal> existingSignal = tradingSignalService.getSignalById(signalId);

            if (existingSignal.isEmpty()) {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("Trading signal not found")
                        .asRuntimeException());
                return;
            }

            tradingSignalService.deleteSignal(signalId);

            DeleteSignalResponse response = DeleteSignalResponse.newBuilder()
                    .setSuccess(true)
                    .setMessage("Trading signal deleted successfully")
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error deleting trading signal", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error deleting trading signal: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getSignalsByDateRange(GetSignalsByDateRangeRequest request, StreamObserver<SignalListResponse> responseObserver) {
        try {
            LocalDateTime startDate = LocalDateTime.ofEpochSecond(
                    request.getStartDate().getSeconds(),
                    request.getStartDate().getNanos(),
                    ZoneOffset.UTC
            );
            LocalDateTime endDate = LocalDateTime.ofEpochSecond(
                    request.getEndDate().getSeconds(),
                    request.getEndDate().getNanos(),
                    ZoneOffset.UTC
            );

            List<TradingSignal> signals = tradingSignalService.getSignalsByDateRange(startDate, endDate);
            List<SignalResponse> signalResponses = signals.stream()
                    .map(this::mapToSignalResponse)
                    .collect(Collectors.toList());

            SignalListResponse response = SignalListResponse.newBuilder()
                    .addAllSignals(signalResponses)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting signals by date range", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting signals by date range: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    private SignalResponse mapToSignalResponse(TradingSignal signal) {
        SignalResponse.Builder builder = SignalResponse.newBuilder()
                .setId(signal.getId().toString())
                .setSymbol(signal.getSymbol())
                .setSignalType(signal.getSignalType())
                .setDirection(signal.getDirection())
                .setStatus(signal.getStatus());

        if (signal.getPredictionId() != null) {
            builder.setPredictionId(signal.getPredictionId().toString());
        }

        if (signal.getStrength() != null) {
            builder.setStrength(signal.getStrength().doubleValue());
        }

        if (signal.getConfidence() != null) {
            builder.setConfidence(signal.getConfidence().doubleValue());
        }

        if (signal.getGeneratedAt() != null) {
            builder.setGeneratedAt(Timestamp.newBuilder()
                    .setSeconds(signal.getGeneratedAt().toEpochSecond(ZoneOffset.UTC))
                    .setNanos(signal.getGeneratedAt().getNano())
                    .build());
        }

        if (signal.getActualReturnPercent() != null) {
            builder.setActualReturnPercent(signal.getActualReturnPercent().doubleValue());
        }

        return builder.build();
    }
}