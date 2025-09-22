package com.alert_signals.AlertSignal.grpc;

import com.alert_signals.AlertSignal.entity.AlertDeliveryLog;
import com.alert_signals.AlertSignal.service.AlertDeliveryService;
import com.alert_signals.AlertSignal.grpc.generated.*;
import com.google.protobuf.Timestamp;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.grpc.server.service.GrpcService;

import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

@GrpcService
@RequiredArgsConstructor
@Slf4j
public class AlertDeliveryGrpcServiceImpl extends AlertDeliveryGrpcServiceGrpc.AlertDeliveryGrpcServiceImplBase {

    private final AlertDeliveryService alertDeliveryService;

    @Override
    public void logDelivery(LogDeliveryRequest request, StreamObserver<DeliveryLogResponse> responseObserver) {
        try {
            AlertDeliveryLog deliveryLog = AlertDeliveryLog.builder()
                    .signalId(UUID.fromString(request.getSignalId()))
                    .subscriptionId(UUID.fromString(request.getSubscriptionId()))
                    .deliveryMethod(request.getDeliveryMethod())
                    .status(request.getStatus())
                    .deliveryLatencyMs(request.getDeliveryLatencyMs())
                    .build();

            AlertDeliveryLog savedDeliveryLog = alertDeliveryService.logDelivery(deliveryLog);
            DeliveryLogResponse response = mapToDeliveryLogResponse(savedDeliveryLog);

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error logging delivery", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error logging delivery: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getDeliveryLog(GetDeliveryLogRequest request, StreamObserver<DeliveryLogResponse> responseObserver) {
        try {
            UUID deliveryLogId = UUID.fromString(request.getId());
            Optional<AlertDeliveryLog> deliveryLog = alertDeliveryService.getDeliveryLogById(deliveryLogId);

            if (deliveryLog.isPresent()) {
                DeliveryLogResponse response = mapToDeliveryLogResponse(deliveryLog.get());
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("Delivery log not found")
                        .asRuntimeException());
            }
        } catch (Exception e) {
            log.error("Error getting delivery log", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting delivery log: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getDeliveryLogsBySignalId(GetDeliveryLogsBySignalIdRequest request, StreamObserver<DeliveryLogListResponse> responseObserver) {
        try {
            UUID signalId = UUID.fromString(request.getSignalId());
            List<AlertDeliveryLog> deliveryLogs = alertDeliveryService.getDeliveryLogsBySignalId(signalId);
            List<DeliveryLogResponse> deliveryLogResponses = deliveryLogs.stream()
                    .map(this::mapToDeliveryLogResponse)
                    .collect(Collectors.toList());

            DeliveryLogListResponse response = DeliveryLogListResponse.newBuilder()
                    .addAllDeliveryLogs(deliveryLogResponses)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting delivery logs by signal ID", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting delivery logs by signal ID: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getDeliveryLogsBySubscriptionId(GetDeliveryLogsBySubscriptionIdRequest request, StreamObserver<DeliveryLogListResponse> responseObserver) {
        try {
            UUID subscriptionId = UUID.fromString(request.getSubscriptionId());
            List<AlertDeliveryLog> deliveryLogs = alertDeliveryService.getDeliveryLogsBySubscriptionId(subscriptionId);
            List<DeliveryLogResponse> deliveryLogResponses = deliveryLogs.stream()
                    .map(this::mapToDeliveryLogResponse)
                    .collect(Collectors.toList());

            DeliveryLogListResponse response = DeliveryLogListResponse.newBuilder()
                    .addAllDeliveryLogs(deliveryLogResponses)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting delivery logs by subscription ID", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting delivery logs by subscription ID: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getDeliveryLogsByStatus(GetDeliveryLogsByStatusRequest request, StreamObserver<DeliveryLogListResponse> responseObserver) {
        try {
            List<AlertDeliveryLog> deliveryLogs = alertDeliveryService.getDeliveryLogsByStatus(request.getStatus());
            List<DeliveryLogResponse> deliveryLogResponses = deliveryLogs.stream()
                    .map(this::mapToDeliveryLogResponse)
                    .collect(Collectors.toList());

            DeliveryLogListResponse response = DeliveryLogListResponse.newBuilder()
                    .addAllDeliveryLogs(deliveryLogResponses)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting delivery logs by status", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting delivery logs by status: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getDeliveryLogsByMethod(GetDeliveryLogsByMethodRequest request, StreamObserver<DeliveryLogListResponse> responseObserver) {
        try {
            List<AlertDeliveryLog> deliveryLogs = alertDeliveryService.getDeliveryLogsByMethod(request.getDeliveryMethod());
            List<DeliveryLogResponse> deliveryLogResponses = deliveryLogs.stream()
                    .map(this::mapToDeliveryLogResponse)
                    .collect(Collectors.toList());

            DeliveryLogListResponse response = DeliveryLogListResponse.newBuilder()
                    .addAllDeliveryLogs(deliveryLogResponses)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting delivery logs by method", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting delivery logs by method: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getDeliveryLogsByDateRange(GetDeliveryLogsByDateRangeRequest request, StreamObserver<DeliveryLogListResponse> responseObserver) {
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

            List<AlertDeliveryLog> deliveryLogs = alertDeliveryService.getDeliveryLogsByDateRange(startDate, endDate);
            List<DeliveryLogResponse> deliveryLogResponses = deliveryLogs.stream()
                    .map(this::mapToDeliveryLogResponse)
                    .collect(Collectors.toList());

            DeliveryLogListResponse response = DeliveryLogListResponse.newBuilder()
                    .addAllDeliveryLogs(deliveryLogResponses)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting delivery logs by date range", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting delivery logs by date range: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getAverageDeliveryLatency(GetAverageDeliveryLatencyRequest request, StreamObserver<AverageLatencyResponse> responseObserver) {
        try {
            Double averageLatency = alertDeliveryService.getAverageDeliveryLatency(request.getDeliveryMethod());

            AverageLatencyResponse response = AverageLatencyResponse.newBuilder()
                    .setAverageLatency(averageLatency != null ? averageLatency : 0.0)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting average delivery latency", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting average delivery latency: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getSuccessfulDeliveryCount(GetSuccessfulDeliveryCountRequest request, StreamObserver<DeliveryCountResponse> responseObserver) {
        try {
            UUID subscriptionId = UUID.fromString(request.getSubscriptionId());
            Long count = alertDeliveryService.getSuccessfulDeliveryCount(subscriptionId);

            DeliveryCountResponse response = DeliveryCountResponse.newBuilder()
                    .setCount(count != null ? count : 0L)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting successful delivery count", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting successful delivery count: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void updateDeliveryLog(UpdateDeliveryLogRequest request, StreamObserver<DeliveryLogResponse> responseObserver) {
        try {
            UUID deliveryLogId = UUID.fromString(request.getId());
            Optional<AlertDeliveryLog> existingDeliveryLog = alertDeliveryService.getDeliveryLogById(deliveryLogId);

            if (existingDeliveryLog.isEmpty()) {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("Delivery log not found")
                        .asRuntimeException());
                return;
            }

            LocalDateTime sentAt = LocalDateTime.ofEpochSecond(
                    request.getSentAt().getSeconds(),
                    request.getSentAt().getNanos(),
                    ZoneOffset.UTC
            );

            AlertDeliveryLog deliveryLog = AlertDeliveryLog.builder()
                    .id(deliveryLogId)
                    .signalId(UUID.fromString(request.getSignalId()))
                    .subscriptionId(UUID.fromString(request.getSubscriptionId()))
                    .deliveryMethod(request.getDeliveryMethod())
                    .status(request.getStatus())
                    .sentAt(sentAt)
                    .deliveryLatencyMs(request.getDeliveryLatencyMs())
                    .build();

            AlertDeliveryLog updatedDeliveryLog = alertDeliveryService.updateDeliveryLog(deliveryLog);
            DeliveryLogResponse response = mapToDeliveryLogResponse(updatedDeliveryLog);

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error updating delivery log", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error updating delivery log: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    private DeliveryLogResponse mapToDeliveryLogResponse(AlertDeliveryLog deliveryLog) {
        DeliveryLogResponse.Builder builder = DeliveryLogResponse.newBuilder()
                .setId(deliveryLog.getId().toString())
                .setSignalId(deliveryLog.getSignalId().toString())
                .setSubscriptionId(deliveryLog.getSubscriptionId().toString())
                .setDeliveryMethod(deliveryLog.getDeliveryMethod())
                .setStatus(deliveryLog.getStatus());

        if (deliveryLog.getSentAt() != null) {
            builder.setSentAt(Timestamp.newBuilder()
                    .setSeconds(deliveryLog.getSentAt().toEpochSecond(ZoneOffset.UTC))
                    .setNanos(deliveryLog.getSentAt().getNano())
                    .build());
        }

        if (deliveryLog.getDeliveryLatencyMs() != null) {
            builder.setDeliveryLatencyMs(deliveryLog.getDeliveryLatencyMs());
        }

        return builder.build();
    }
}