package com.alert_signals.AlertSignal.repository;

import com.alert_signals.AlertSignal.entity.AlertDeliveryLog;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

@Repository
public interface AlertDeliveryLogRepository extends JpaRepository<AlertDeliveryLog, UUID> {

    List<AlertDeliveryLog> findBySignalId(UUID signalId);

    List<AlertDeliveryLog> findBySubscriptionId(UUID subscriptionId);

    List<AlertDeliveryLog> findByStatus(String status);

    List<AlertDeliveryLog> findByDeliveryMethod(String deliveryMethod);

    @Query("SELECT adl FROM AlertDeliveryLog adl WHERE adl.sentAt BETWEEN :startDate AND :endDate")
    List<AlertDeliveryLog> findBySentAtBetween(@Param("startDate") LocalDateTime startDate,
                                               @Param("endDate") LocalDateTime endDate);

    @Query("SELECT AVG(adl.deliveryLatencyMs) FROM AlertDeliveryLog adl WHERE adl.deliveryMethod = :deliveryMethod AND adl.status = 'SUCCESS'")
    Double getAverageDeliveryLatencyByMethod(@Param("deliveryMethod") String deliveryMethod);

    @Query("SELECT COUNT(adl) FROM AlertDeliveryLog adl WHERE adl.subscriptionId = :subscriptionId AND adl.status = 'SUCCESS'")
    Long countSuccessfulDeliveriesBySubscription(@Param("subscriptionId") UUID subscriptionId);
}
