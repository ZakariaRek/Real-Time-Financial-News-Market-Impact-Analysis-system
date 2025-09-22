package com.alert_signals.AlertSignal.grpc;

import com.alert_signals.AlertSignal.entity.UserSubscription;
import com.alert_signals.AlertSignal.service.UserSubscriptionService;
import com.alert_signals.AlertSignal.grpc.generated.*;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.grpc.server.service.GrpcService;

import java.math.BigDecimal;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

@GrpcService
@RequiredArgsConstructor
@Slf4j
public class UserSubscriptionGrpcServiceImpl extends UserSubscriptionGrpcServiceGrpc.UserSubscriptionGrpcServiceImplBase {

    private final UserSubscriptionService userSubscriptionService;

    @Override
    public void createSubscription(CreateSubscriptionRequest request, StreamObserver<SubscriptionResponse> responseObserver) {
        try {
            UserSubscription subscription = UserSubscription.builder()
                    .userId(request.getUserId())
                    .subscriptionName(request.getSubscriptionName())
                    .symbols(request.getSymbolsList())
                    .signalTypes(request.getSignalTypesList())
                    .minConfidence(BigDecimal.valueOf(request.getMinConfidence()))
                    .deliveryMethods(request.getDeliveryMethodsList())
                    .isActive(request.getIsActive())
                    .build();

            UserSubscription savedSubscription = userSubscriptionService.createSubscription(subscription);
            SubscriptionResponse response = mapToSubscriptionResponse(savedSubscription);

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error creating subscription", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error creating subscription: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getSubscription(GetSubscriptionRequest request, StreamObserver<SubscriptionResponse> responseObserver) {
        try {
            UUID subscriptionId = UUID.fromString(request.getId());
            Optional<UserSubscription> subscription = userSubscriptionService.getSubscriptionById(subscriptionId);

            if (subscription.isPresent()) {
                SubscriptionResponse response = mapToSubscriptionResponse(subscription.get());
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("Subscription not found")
                        .asRuntimeException());
            }
        } catch (Exception e) {
            log.error("Error getting subscription", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting subscription: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getSubscriptionByName(GetSubscriptionByNameRequest request, StreamObserver<SubscriptionResponse> responseObserver) {
        try {
            Optional<UserSubscription> subscription = userSubscriptionService.getSubscriptionByName(request.getSubscriptionName());

            if (subscription.isPresent()) {
                SubscriptionResponse response = mapToSubscriptionResponse(subscription.get());
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("Subscription not found")
                        .asRuntimeException());
            }
        } catch (Exception e) {
            log.error("Error getting subscription by name", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting subscription by name: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getSubscriptionsByUserId(GetSubscriptionsByUserIdRequest request, StreamObserver<SubscriptionListResponse> responseObserver) {
        try {
            List<UserSubscription> subscriptions = userSubscriptionService.getSubscriptionsByUserId(request.getUserId());
            List<SubscriptionResponse> subscriptionResponses = subscriptions.stream()
                    .map(this::mapToSubscriptionResponse)
                    .collect(Collectors.toList());

            SubscriptionListResponse response = SubscriptionListResponse.newBuilder()
                    .addAllSubscriptions(subscriptionResponses)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting subscriptions by user ID", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting subscriptions by user ID: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getActiveSubscriptions(GetActiveSubscriptionsRequest request, StreamObserver<SubscriptionListResponse> responseObserver) {
        try {
            List<UserSubscription> subscriptions = userSubscriptionService.getActiveSubscriptions();
            List<SubscriptionResponse> subscriptionResponses = subscriptions.stream()
                    .map(this::mapToSubscriptionResponse)
                    .collect(Collectors.toList());

            SubscriptionListResponse response = SubscriptionListResponse.newBuilder()
                    .addAllSubscriptions(subscriptionResponses)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting active subscriptions", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting active subscriptions: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getActiveSubscriptionsBySymbol(GetActiveSubscriptionsBySymbolRequest request, StreamObserver<SubscriptionListResponse> responseObserver) {
        try {
            List<UserSubscription> subscriptions = userSubscriptionService.getActiveSubscriptionsBySymbol(request.getSymbol());
            List<SubscriptionResponse> subscriptionResponses = subscriptions.stream()
                    .map(this::mapToSubscriptionResponse)
                    .collect(Collectors.toList());

            SubscriptionListResponse response = SubscriptionListResponse.newBuilder()
                    .addAllSubscriptions(subscriptionResponses)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting active subscriptions by symbol", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting active subscriptions by symbol: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void getActiveSubscriptionsBySignalType(GetActiveSubscriptionsBySignalTypeRequest request, StreamObserver<SubscriptionListResponse> responseObserver) {
        try {
            List<UserSubscription> subscriptions = userSubscriptionService.getActiveSubscriptionsBySignalType(request.getSignalType());
            List<SubscriptionResponse> subscriptionResponses = subscriptions.stream()
                    .map(this::mapToSubscriptionResponse)
                    .collect(Collectors.toList());

            SubscriptionListResponse response = SubscriptionListResponse.newBuilder()
                    .addAllSubscriptions(subscriptionResponses)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting active subscriptions by signal type", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error getting active subscriptions by signal type: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void updateSubscription(UpdateSubscriptionRequest request, StreamObserver<SubscriptionResponse> responseObserver) {
        try {
            UUID subscriptionId = UUID.fromString(request.getId());
            Optional<UserSubscription> existingSubscription = userSubscriptionService.getSubscriptionById(subscriptionId);

            if (existingSubscription.isEmpty()) {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("Subscription not found")
                        .asRuntimeException());
                return;
            }

            UserSubscription subscription = UserSubscription.builder()
                    .id(subscriptionId)
                    .userId(request.getUserId())
                    .subscriptionName(request.getSubscriptionName())
                    .symbols(request.getSymbolsList())
                    .signalTypes(request.getSignalTypesList())
                    .minConfidence(BigDecimal.valueOf(request.getMinConfidence()))
                    .deliveryMethods(request.getDeliveryMethodsList())
                    .isActive(request.getIsActive())
                    .build();

            UserSubscription updatedSubscription = userSubscriptionService.updateSubscription(subscription);
            SubscriptionResponse response = mapToSubscriptionResponse(updatedSubscription);

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error updating subscription", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error updating subscription: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void deactivateSubscription(DeactivateSubscriptionRequest request, StreamObserver<DeactivateSubscriptionResponse> responseObserver) {
        try {
            UUID subscriptionId = UUID.fromString(request.getId());
            Optional<UserSubscription> existingSubscription = userSubscriptionService.getSubscriptionById(subscriptionId);

            if (existingSubscription.isEmpty()) {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("Subscription not found")
                        .asRuntimeException());
                return;
            }

            userSubscriptionService.deactivateSubscription(subscriptionId);

            DeactivateSubscriptionResponse response = DeactivateSubscriptionResponse.newBuilder()
                    .setSuccess(true)
                    .setMessage("Subscription deactivated successfully")
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error deactivating subscription", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error deactivating subscription: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    @Override
    public void deleteSubscription(DeleteSubscriptionRequest request, StreamObserver<DeleteSubscriptionResponse> responseObserver) {
        try {
            UUID subscriptionId = UUID.fromString(request.getId());
            Optional<UserSubscription> existingSubscription = userSubscriptionService.getSubscriptionById(subscriptionId);

            if (existingSubscription.isEmpty()) {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("Subscription not found")
                        .asRuntimeException());
                return;
            }

            userSubscriptionService.deleteSubscription(subscriptionId);

            DeleteSubscriptionResponse response = DeleteSubscriptionResponse.newBuilder()
                    .setSuccess(true)
                    .setMessage("Subscription deleted successfully")
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error deleting subscription", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription("Error deleting subscription: " + e.getMessage())
                    .asRuntimeException());
        }
    }

    private SubscriptionResponse mapToSubscriptionResponse(UserSubscription subscription) {
        SubscriptionResponse.Builder builder = SubscriptionResponse.newBuilder()
                .setId(subscription.getId().toString())
                .setUserId(subscription.getUserId())
                .setSubscriptionName(subscription.getSubscriptionName())
                .addAllSymbols(subscription.getSymbols() != null ? subscription.getSymbols() : List.of())
                .addAllSignalTypes(subscription.getSignalTypes() != null ? subscription.getSignalTypes() : List.of())
                .addAllDeliveryMethods(subscription.getDeliveryMethods() != null ? subscription.getDeliveryMethods() : List.of())
                .setIsActive(subscription.getIsActive());

        if (subscription.getMinConfidence() != null) {
            builder.setMinConfidence(subscription.getMinConfidence().doubleValue());
        }

        return builder.build();
    }
}