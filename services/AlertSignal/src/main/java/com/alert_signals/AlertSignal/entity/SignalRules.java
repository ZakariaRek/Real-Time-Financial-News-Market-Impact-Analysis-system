package com.alert_signals.AlertSignal.entity;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.GenericGenerator;


import java.math.BigDecimal;
import java.util.List;
import java.util.UUID;

@Entity
@Table(name = "signal_rules")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class SignalRules {

    @Id
    @GeneratedValue(generator = "UUID")
    @GenericGenerator(name = "UUID", strategy = "org.hibernate.id.UUIDGenerator")
    @Column(name = "id", updatable = false, nullable = false)
    private UUID id;

    @Column(name = "rule_name", unique = true, nullable = false)
    private String ruleName;

    @Column(name = "rule_type", nullable = false)
    private String ruleType;

    @Column(name = "conditions", columnDefinition = "jsonb")
    private String conditions;

    @Column(name = "symbols", columnDefinition = "text[]")
    private List<String> symbols;

    @Column(name = "success_rate", precision = 10, scale = 4)
    private BigDecimal successRate;

    @Column(name = "is_active", nullable = false)
    private Boolean isActive;
}