package com.alert_signals.AlertSignal.controller;

import com.alert_signals.AlertSignal.dto.UserSubscriptionDto;
import com.alert_signals.AlertSignal.entity.UserSubscription;
import com.alert_signals.AlertSignal.mapper.UserSubscriptionMapper;
import com.alert_signals.AlertSignal.service.UserSubscriptionService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@RestController
@RequestMapping("/user-subscriptions")
@RequiredArgsConstructor
public class UserSubscriptionController {

    private final UserSubscriptionService userSubscriptionService;
    private final UserSubscriptionMapper userSubscriptionMapper;

    @PostMapping
    public ResponseEntity<UserSubscriptionDto> createSubscription(@RequestBody UserSubscriptionDto subscriptionDto) {
        UserSubscription subscription = userSubscriptionMapper.toEntity(subscriptionDto);
        UserSubscription savedSubscription = userSubscriptionService.createSubscription(subscription);
        UserSubscriptionDto responseDto = userSubscriptionMapper.toDto(savedSubscription);
        return new ResponseEntity<>(responseDto, HttpStatus.CREATED);
    }

    @GetMapping("/{id}")
    public ResponseEntity<UserSubscriptionDto> getSubscriptionById(@PathVariable UUID id) {
        Optional<UserSubscription> subscription = userSubscriptionService.getSubscriptionById(id);
        return subscription.map(s -> ResponseEntity.ok(userSubscriptionMapper.toDto(s)))
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/name/{subscriptionName}")
    public ResponseEntity<UserSubscriptionDto> getSubscriptionByName(@PathVariable String subscriptionName) {
        Optional<UserSubscription> subscription = userSubscriptionService.getSubscriptionByName(subscriptionName);
        return subscription.map(s -> ResponseEntity.ok(userSubscriptionMapper.toDto(s)))
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/user/{userId}")
    public ResponseEntity<List<UserSubscriptionDto>> getSubscriptionsByUserId(@PathVariable String userId) {
        List<UserSubscription> subscriptions = userSubscriptionService.getSubscriptionsByUserId(userId);
        List<UserSubscriptionDto> subscriptionDtos = userSubscriptionMapper.toDtoList(subscriptions);
        return ResponseEntity.ok(subscriptionDtos);
    }

    @GetMapping("/active")
    public ResponseEntity<List<UserSubscriptionDto>> getActiveSubscriptions() {
        List<UserSubscription> subscriptions = userSubscriptionService.getActiveSubscriptions();
        List<UserSubscriptionDto> subscriptionDtos = userSubscriptionMapper.toDtoList(subscriptions);
        return ResponseEntity.ok(subscriptionDtos);
    }

    @GetMapping("/symbol/{symbol}")
    public ResponseEntity<List<UserSubscriptionDto>> getActiveSubscriptionsBySymbol(@PathVariable String symbol) {
        List<UserSubscription> subscriptions = userSubscriptionService.getActiveSubscriptionsBySymbol(symbol);
        List<UserSubscriptionDto> subscriptionDtos = userSubscriptionMapper.toDtoList(subscriptions);
        return ResponseEntity.ok(subscriptionDtos);
    }

    @GetMapping("/signal-type/{signalType}")
    public ResponseEntity<List<UserSubscriptionDto>> getActiveSubscriptionsBySignalType(@PathVariable String signalType) {
        List<UserSubscription> subscriptions = userSubscriptionService.getActiveSubscriptionsBySignalType(signalType);
        List<UserSubscriptionDto> subscriptionDtos = userSubscriptionMapper.toDtoList(subscriptions);
        return ResponseEntity.ok(subscriptionDtos);
    }

    @PutMapping("/{id}")
    public ResponseEntity<UserSubscriptionDto> updateSubscription(
            @PathVariable UUID id,
            @RequestBody UserSubscriptionDto subscriptionDto) {
        Optional<UserSubscription> existingSubscription = userSubscriptionService.getSubscriptionById(id);
        if (existingSubscription.isEmpty()) {
            return ResponseEntity.notFound().build();
        }

        subscriptionDto.setId(id);
        UserSubscription subscription = userSubscriptionMapper.toEntity(subscriptionDto);
        UserSubscription updatedSubscription = userSubscriptionService.updateSubscription(subscription);
        UserSubscriptionDto responseDto = userSubscriptionMapper.toDto(updatedSubscription);
        return ResponseEntity.ok(responseDto);
    }

    @PatchMapping("/{id}/deactivate")
    public ResponseEntity<Void> deactivateSubscription(@PathVariable UUID id) {
        Optional<UserSubscription> existingSubscription = userSubscriptionService.getSubscriptionById(id);
        if (existingSubscription.isEmpty()) {
            return ResponseEntity.notFound().build();
        }

        userSubscriptionService.deactivateSubscription(id);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteSubscription(@PathVariable UUID id) {
        Optional<UserSubscription> existingSubscription = userSubscriptionService.getSubscriptionById(id);
        if (existingSubscription.isEmpty()) {
            return ResponseEntity.notFound().build();
        }

        userSubscriptionService.deleteSubscription(id);
        return ResponseEntity.noContent().build();
    }
}