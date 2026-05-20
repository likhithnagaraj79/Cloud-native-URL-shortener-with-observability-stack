package com.urlshortener.metrics;

import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.MeterRegistry;
import org.springframework.stereotype.Component;

@Component
public class UrlMetrics {

    private final Counter urlsCreated;
    private final Counter cacheHits;
    private final Counter cacheMisses;
    private final MeterRegistry registry;

    public UrlMetrics(MeterRegistry registry) {
        this.registry = registry;
        this.urlsCreated = Counter.builder("url_shortener_urls_created_total")
                .description("Total short URLs created")
                .register(registry);
        this.cacheHits = Counter.builder("url_shortener_cache_hits_total")
                .description("Redis cache hits")
                .register(registry);
        this.cacheMisses = Counter.builder("url_shortener_cache_misses_total")
                .description("Redis cache misses")
                .register(registry);
    }

    public void urlCreated()           { urlsCreated.increment(); }
    public void cacheHit()             { cacheHits.increment(); }
    public void cacheMiss()            { cacheMisses.increment(); }

    public void redirect(String code) {
        registry.counter("url_shortener_redirects_total", "short_code", code).increment();
    }
}
