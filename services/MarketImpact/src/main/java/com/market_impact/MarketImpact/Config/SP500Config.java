package com.market_impact.MarketImpact.Config;

import lombok.Getter;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

import java.util.Arrays;
import java.util.List;

@Configuration
@Getter
public class SP500Config {

    // Top S&P 500 stocks by market cap
    private List<String> symbols = Arrays.asList(
            "AAPL", "MSFT", "GOOGL", "AMZN", "NVDA", "META", "TSLA", "BRK.B", "UNH", "XOM",
            "JNJ", "JPM", "V", "PG", "MA", "HD", "CVX", "MRK", "ABBV", "PEP",
            "KO", "AVGO", "COST", "WMT", "MCD", "CSCO", "ACN", "LIN", "TMO", "ABT",
            "ADBE", "NFLX", "NKE", "DHR", "VZ", "TXN", "PM", "CRM", "NEE", "CMCSA",
            "INTC", "DIS", "AMD", "QCOM", "RTX", "INTU", "HON", "UPS", "BA", "AMGN"
    );

    private int batchSize = 10;
    private int maxConcurrent = 5;
}