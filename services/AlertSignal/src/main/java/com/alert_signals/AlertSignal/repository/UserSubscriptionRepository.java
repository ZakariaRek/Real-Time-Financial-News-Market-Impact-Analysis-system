package com.alert_signals.AlertSignal.repository;

import com.alert_signals.AlertSignal.entity.UserSubscription;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface UserSubscriptionRepository extends JpaRepository<UserSubscription, UUID> {

    List<UserSubscription> findByUserId(String userId);

    Optional<UserSubscription> findBySubscriptionName(String subscriptionName);

    List<UserSubscription> findByIsActive(Boolean isActive);

    List<UserSubscription> findByUserIdAndIsActive(String userId, Boolean isActive);

    @Query("SELECT us FROM UserSubscription us WHERE :symbol = ANY(us.symbols) AND us.isActive = true")
    List<UserSubscription> findActiveSubscriptionsBySymbol(@Param("symbol") String symbol);

    @Query("SELECT us FROM UserSubscription us WHERE :signalType = ANY(us.signalTypes) AND us.isActive = true")
    List<UserSubscription> findActiveSubscriptionsBySignalType(@Param("signalType") String signalType);
}
