package com.alert_signals.AlertSignal.repository;

import com.alert_signals.AlertSignal.entity.SignalRules;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface SignalRulesRepository extends JpaRepository<SignalRules, UUID> {

    Optional<SignalRules> findByRuleName(String ruleName);

    List<SignalRules> findByRuleType(String ruleType);

    List<SignalRules> findByIsActive(Boolean isActive);

    List<SignalRules> findByRuleTypeAndIsActive(String ruleType, Boolean isActive);

    @Query("SELECT sr FROM SignalRules sr WHERE :symbol = ANY(sr.symbols) AND sr.isActive = true")
    List<SignalRules> findActiveRulesBySymbol(@Param("symbol") String symbol);

    @Query("SELECT sr FROM SignalRules sr WHERE sr.successRate >= :minSuccessRate AND sr.isActive = true")
    List<SignalRules> findActiveRulesByMinSuccessRate(@Param("minSuccessRate") Double minSuccessRate);
}