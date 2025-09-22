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

    // Fixed: Using native SQL query for PostgreSQL array operations
    @Query(value = "SELECT * FROM user_subscriptions us WHERE :symbol = ANY(us.symbols) AND us.is_active = true",
            nativeQuery = true)
    List<UserSubscription> findActiveSubscriptionsBySymbol(@Param("symbol") String symbol);

    // Fixed: Using native SQL query for PostgreSQL array operations
    @Query(value = "SELECT * FROM user_subscriptions us WHERE :signalType = ANY(us.signal_types) AND us.is_active = true",
            nativeQuery = true)
    List<UserSubscription> findActiveSubscriptionsBySignalType(@Param("signalType") String signalType);
}