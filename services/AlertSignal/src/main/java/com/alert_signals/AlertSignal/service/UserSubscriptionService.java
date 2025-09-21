package com.alert_signals.AlertSignal.service;

import com.alert_signals.AlertSignal.entity.UserSubscription;
import com.alert_signals.AlertSignal.repository.UserSubscriptionRepository;
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
public class UserSubscriptionService {

    private final UserSubscriptionRepository userSubscriptionRepository;

    public UserSubscription createSubscription(UserSubscription subscription) {
        log.info("Creating new subscription: {} for user: {}",
                subscription.getSubscriptionName(), subscription.getUserId());
        return userSubscriptionRepository.save(subscription);
    }

    public Optional<UserSubscription> getSubscriptionById(UUID id) {
        return userSubscriptionRepository.findById(id);
    }

    public Optional<UserSubscription> getSubscriptionByName(String subscriptionName) {
        return userSubscriptionRepository.findBySubscriptionName(subscriptionName);
    }

    public List<UserSubscription> getSubscriptionsByUserId(String userId) {
        return userSubscriptionRepository.findByUserId(userId);
    }

    public List<UserSubscription> getActiveSubscriptions() {
        return userSubscriptionRepository.findByIsActive(true);
    }

    public List<UserSubscription> getActiveSubscriptionsBySymbol(String symbol) {
        return userSubscriptionRepository.findActiveSubscriptionsBySymbol(symbol);
    }

    public List<UserSubscription> getActiveSubscriptionsBySignalType(String signalType) {
        return userSubscriptionRepository.findActiveSubscriptionsBySignalType(signalType);
    }

    public UserSubscription updateSubscription(UserSubscription subscription) {
        log.info("Updating subscription with ID: {}", subscription.getId());
        return userSubscriptionRepository.save(subscription);
    }

    public void deactivateSubscription(UUID id) {
        log.info("Deactivating subscription with ID: {}", id);
        Optional<UserSubscription> subscription = userSubscriptionRepository.findById(id);
        if (subscription.isPresent()) {
            subscription.get().setIsActive(false);
            userSubscriptionRepository.save(subscription.get());
        }
    }

    public void deleteSubscription(UUID id) {
        log.info("Deleting subscription with ID: {}", id);
        userSubscriptionRepository.deleteById(id);
    }
}