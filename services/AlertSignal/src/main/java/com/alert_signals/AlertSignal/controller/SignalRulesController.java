package com.alert_signals.AlertSignal.controller;

import com.alert_signals.AlertSignal.dto.SignalRulesDto;
import com.alert_signals.AlertSignal.entity.SignalRules;
import com.alert_signals.AlertSignal.mapper.SignalRulesMapper;
import com.alert_signals.AlertSignal.service.SignalRulesService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@RestController
@RequestMapping("/signal-rules")
@RequiredArgsConstructor
public class SignalRulesController {

    private final SignalRulesService signalRulesService;
    private final SignalRulesMapper signalRulesMapper;

    @PostMapping
    public ResponseEntity<SignalRulesDto> createRule(@RequestBody SignalRulesDto ruleDto) {
        SignalRules rule = signalRulesMapper.toEntity(ruleDto);
        SignalRules savedRule = signalRulesService.createRule(rule);
        SignalRulesDto responseDto = signalRulesMapper.toDto(savedRule);
        return new ResponseEntity<>(responseDto, HttpStatus.CREATED);
    }

    @GetMapping("/{id}")
    public ResponseEntity<SignalRulesDto> getRuleById(@PathVariable UUID id) {
        Optional<SignalRules> rule = signalRulesService.getRuleById(id);
        return rule.map(r -> ResponseEntity.ok(signalRulesMapper.toDto(r)))
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/name/{ruleName}")
    public ResponseEntity<SignalRulesDto> getRuleByName(@PathVariable String ruleName) {
        Optional<SignalRules> rule = signalRulesService.getRuleByName(ruleName);
        return rule.map(r -> ResponseEntity.ok(signalRulesMapper.toDto(r)))
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping
    public ResponseEntity<List<SignalRulesDto>> getAllRules() {
        List<SignalRules> rules = signalRulesService.getAllRules();
        List<SignalRulesDto> ruleDtos = signalRulesMapper.toDtoList(rules);
        return ResponseEntity.ok(ruleDtos);
    }

    @GetMapping("/type/{ruleType}")
    public ResponseEntity<List<SignalRulesDto>> getRulesByType(@PathVariable String ruleType) {
        List<SignalRules> rules = signalRulesService.getRulesByType(ruleType);
        List<SignalRulesDto> ruleDtos = signalRulesMapper.toDtoList(rules);
        return ResponseEntity.ok(ruleDtos);
    }

    @GetMapping("/active")
    public ResponseEntity<List<SignalRulesDto>> getActiveRules() {
        List<SignalRules> rules = signalRulesService.getActiveRules();
        List<SignalRulesDto> ruleDtos = signalRulesMapper.toDtoList(rules);
        return ResponseEntity.ok(ruleDtos);
    }

    @GetMapping("/active/type/{ruleType}")
    public ResponseEntity<List<SignalRulesDto>> getActiveRulesByType(@PathVariable String ruleType) {
        List<SignalRules> rules = signalRulesService.getActiveRulesByType(ruleType);
        List<SignalRulesDto> ruleDtos = signalRulesMapper.toDtoList(rules);
        return ResponseEntity.ok(ruleDtos);
    }

    @GetMapping("/active/symbol/{symbol}")
    public ResponseEntity<List<SignalRulesDto>> getActiveRulesBySymbol(@PathVariable String symbol) {
        List<SignalRules> rules = signalRulesService.getActiveRulesBySymbol(symbol);
        List<SignalRulesDto> ruleDtos = signalRulesMapper.toDtoList(rules);
        return ResponseEntity.ok(ruleDtos);
    }

    @GetMapping("/active/success-rate/{minSuccessRate}")
    public ResponseEntity<List<SignalRulesDto>> getActiveRulesByMinSuccessRate(@PathVariable Double minSuccessRate) {
        List<SignalRules> rules = signalRulesService.getActiveRulesByMinSuccessRate(minSuccessRate);
        List<SignalRulesDto> ruleDtos = signalRulesMapper.toDtoList(rules);
        return ResponseEntity.ok(ruleDtos);
    }

    @PutMapping("/{id}")
    public ResponseEntity<SignalRulesDto> updateRule(
            @PathVariable UUID id,
            @RequestBody SignalRulesDto ruleDto) {
        Optional<SignalRules> existingRule = signalRulesService.getRuleById(id);
        if (existingRule.isEmpty()) {
            return ResponseEntity.notFound().build();
        }

        ruleDto.setId(id);
        SignalRules rule = signalRulesMapper.toEntity(ruleDto);
        SignalRules updatedRule = signalRulesService.updateRule(rule);
        SignalRulesDto responseDto = signalRulesMapper.toDto(updatedRule);
        return ResponseEntity.ok(responseDto);
    }

    @PatchMapping("/{id}/deactivate")
    public ResponseEntity<Void> deactivateRule(@PathVariable UUID id) {
        Optional<SignalRules> existingRule = signalRulesService.getRuleById(id);
        if (existingRule.isEmpty()) {
            return ResponseEntity.notFound().build();
        }

        signalRulesService.deactivateRule(id);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteRule(@PathVariable UUID id) {
        Optional<SignalRules> existingRule = signalRulesService.getRuleById(id);
        if (existingRule.isEmpty()) {
            return ResponseEntity.notFound().build();
        }

        signalRulesService.deleteRule(id);
        return ResponseEntity.noContent().build();
    }
}