package com.alert_signals.AlertSignal.service;

import com.alert_signals.AlertSignal.entity.SignalRules;
import com.alert_signals.AlertSignal.repository.SignalRulesRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
@Transactional
public class SignalRulesService {

    private final SignalRulesRepository signalRulesRepository;

    public SignalRules createRule(SignalRules rule) {
        log.info("Creating new signal rule: {}", rule.getRuleName());
        return signalRulesRepository.save(rule);
    }

    public Optional<SignalRules> getRuleById(UUID id) {
        return signalRulesRepository.findById(id);
    }

    public Optional<SignalRules> getRuleByName(String ruleName) {
        return signalRulesRepository.findByRuleName(ruleName);
    }

    public List<SignalRules> getRulesByType(String ruleType) {
        return signalRulesRepository.findByRuleType(ruleType);
    }

    public List<SignalRules> getActiveRules() {
        return signalRulesRepository.findByIsActive(true);
    }

    public List<SignalRules> getActiveRulesByType(String ruleType) {
        return signalRulesRepository.findByRuleTypeAndIsActive(ruleType, true);
    }

    public List<SignalRules> getActiveRulesBySymbol(String symbol) {
        return signalRulesRepository.findActiveRulesBySymbol(symbol);
    }

    public List<SignalRules> getActiveRulesByMinSuccessRate(Double minSuccessRate) {
        return signalRulesRepository.findActiveRulesByMinSuccessRate(minSuccessRate);
    }

    public SignalRules updateRule(SignalRules rule) {
        log.info("Updating signal rule with ID: {}", rule.getId());
        return signalRulesRepository.save(rule);
    }

    public void deactivateRule(UUID id) {
        log.info("Deactivating signal rule with ID: {}", id);
        Optional<SignalRules> rule = signalRulesRepository.findById(id);
        if (rule.isPresent()) {
            rule.get().setIsActive(false);
            signalRulesRepository.save(rule.get());
        }
    }

    public void deleteRule(UUID id) {
        log.info("Deleting signal rule with ID: {}", id);
        signalRulesRepository.deleteById(id);
    }

    public List<SignalRules> getAllRules() {
        return signalRulesRepository.findAll();
    }
}