package com.alert_signals.AlertSignal.entity;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.GenericGenerator;
import org.hibernate.annotations.Type;

import java.math.BigDecimal;
import java.util.List;
import java.util.UUID;

@Entity
@Table(name = "user_subscriptions")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class UserSubscription {

    @Id
    @GeneratedValue(generator = "UUID")
    @GenericGenerator(name = "UUID", strategy = "org.hibernate.id.UUIDGenerator")
    @Column(name = "id", updatable = false, nullable = false)
    private UUID id;

    @Column(name = "user_id", nullable = false)
    private String userId;

    @Column(name = "subscription_name", unique = true, nullable = false)
    private String subscriptionName;

    @Column(name = "symbols", columnDefinition = "text[]")
    private List<String> symbols;

    @Column(name = "signal_types", columnDefinition = "text[]")
    private List<String> signalTypes;

    @Column(name = "min_confidence", precision = 10, scale = 4)
    private BigDecimal minConfidence;

    @Column(name = "delivery_methods", columnDefinition = "text[]")
    private List<String> deliveryMethods;

    @Column(name = "is_active", nullable = false)
    private Boolean isActive;
}
