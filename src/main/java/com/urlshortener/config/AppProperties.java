package com.urlshortener.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties("app")
public record AppProperties(
        String baseUrl,
        int shortCodeLength,
        RateLimit rateLimit
) {
    public record RateLimit(int requestsPerMinute) {}
}
