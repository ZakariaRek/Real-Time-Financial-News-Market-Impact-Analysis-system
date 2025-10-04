package com.alert_signals.AlertSignal;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.grpc.client.ImportGrpcClients;


@ImportGrpcClients(basePackageClasses = AlertSignalApplication.class)
@SpringBootApplication
public class AlertSignalApplication {

	public static void main(String[] args) {
		SpringApplication.run(AlertSignalApplication.class, args);
	}

}
