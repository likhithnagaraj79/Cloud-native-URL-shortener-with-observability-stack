package com.urlshortener.service;

import com.urlshortener.config.AppProperties;
import com.urlshortener.dto.CreateUrlRequest;
import com.urlshortener.dto.CreateUrlResponse;
import com.urlshortener.dto.UrlStatsResponse;
import com.urlshortener.exception.CodeAlreadyTakenException;
import com.urlshortener.exception.UrlNotFoundException;
import com.urlshortener.metrics.UrlMetrics;
import com.urlshortener.model.Url;
import com.urlshortener.repository.UrlRepository;
import com.urlshortener.util.ShortCodeGenerator;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Duration;
import java.time.Instant;

@Service
public class UrlServiceImpl implements UrlService {

    private static final Logger log = LoggerFactory.getLogger(UrlServiceImpl.class);
    private static final Duration DEFAULT_CACHE_TTL = Duration.ofHours(24);
    private static final int MAX_RETRIES = 5;

    private final UrlRepository repo;
    private final StringRedisTemplate redis;
    private final ShortCodeGenerator generator;
    private final UrlMetrics metrics;
    private final AppProperties props;

    public UrlServiceImpl(UrlRepository repo, StringRedisTemplate redis,
                          ShortCodeGenerator generator, UrlMetrics metrics,
                          AppProperties props) {
        this.repo = repo;
        this.redis = redis;
        this.generator = generator;
        this.metrics = metrics;
        this.props = props;
    }

    @Override
    @Transactional
    public CreateUrlResponse create(CreateUrlRequest request, String userAgent) {
        String code = request.customCode();
        if (code != null && !code.isBlank()) {
            if (repo.existsByShortCode(code)) {
                throw new CodeAlreadyTakenException(code);
            }
        } else {
            code = generateUniqueCode();
        }

        Url url = new Url();
        url.setShortCode(code);
        url.setOriginalUrl(request.originalUrl());
        url.setExpiresAt(request.expiresAt());
        url.setUserAgent(userAgent);
        repo.save(url);

        cacheUrl(url);
        metrics.urlCreated();

        return new CreateUrlResponse(
                code,
                props.baseUrl() + "/" + code,
                request.originalUrl(),
                url.getCreatedAt(),
                url.getExpiresAt()
        );
    }

    @Override
    public String resolve(String shortCode) {
        String cached = redis.opsForValue().get(cacheKey(shortCode));
        if (cached != null) {
            metrics.cacheHit();
            metrics.redirect(shortCode);
            incrementClickAsync(shortCode);
            return cached;
        }
        metrics.cacheMiss();

        Url url = repo.findActive(shortCode, Instant.now())
                .orElseThrow(() -> new UrlNotFoundException(shortCode));

        cacheUrl(url);
        metrics.redirect(shortCode);
        incrementClickAsync(shortCode);
        return url.getOriginalUrl();
    }

    @Override
    @Transactional(readOnly = true)
    public UrlStatsResponse stats(String shortCode) {
        Url url = repo.findByShortCode(shortCode)
                .orElseThrow(() -> new UrlNotFoundException(shortCode));
        return new UrlStatsResponse(
                url.getShortCode(),
                url.getOriginalUrl(),
                url.getClickCount(),
                url.getCreatedAt(),
                url.getExpiresAt()
        );
    }

    private String generateUniqueCode() {
        for (int i = 0; i < MAX_RETRIES; i++) {
            String code = generator.generate(props.shortCodeLength());
            if (!repo.existsByShortCode(code)) {
                return code;
            }
        }
        throw new IllegalStateException("Failed to generate unique code after " + MAX_RETRIES + " attempts");
    }

    private void cacheUrl(Url url) {
        Duration ttl = DEFAULT_CACHE_TTL;
        if (url.getExpiresAt() != null) {
            Duration remaining = Duration.between(Instant.now(), url.getExpiresAt());
            if (remaining.isNegative() || remaining.isZero()) return;
            if (remaining.compareTo(ttl) < 0) ttl = remaining;
        }
        try {
            redis.opsForValue().set(cacheKey(url.getShortCode()), url.getOriginalUrl(), ttl);
        } catch (Exception e) {
            log.warn("Failed to cache URL {}: {}", url.getShortCode(), e.getMessage());
        }
    }

    @Async
    void incrementClickAsync(String shortCode) {
        try {
            repo.incrementClickCount(shortCode);
        } catch (Exception e) {
            log.warn("Failed to increment click count for {}: {}", shortCode, e.getMessage());
        }
    }

    private static String cacheKey(String code) {
        return "url:" + code;
    }
}
