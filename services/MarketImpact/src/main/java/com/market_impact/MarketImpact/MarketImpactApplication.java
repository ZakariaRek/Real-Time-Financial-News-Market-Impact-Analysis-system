package com.market_impact.MarketImpact;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling // Add this annotation
public class MarketImpactApplication {

	public static void main(String[] args) {
		SpringApplication.run(MarketImpactApplication.class, args);
	}

}
