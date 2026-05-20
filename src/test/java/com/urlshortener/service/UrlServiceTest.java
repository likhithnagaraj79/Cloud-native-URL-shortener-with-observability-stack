package com.urlshortener.service;

import com.urlshortener.config.AppProperties;
import com.urlshortener.dto.CreateUrlRequest;
import com.urlshortener.exception.CodeAlreadyTakenException;
import com.urlshortener.exception.UrlNotFoundException;
import com.urlshortener.metrics.UrlMetrics;
import com.urlshortener.model.Url;
import com.urlshortener.repository.UrlRepository;
import com.urlshortener.util.ShortCodeGenerator;
import io.micrometer.core.instrument.simple.SimpleMeterRegistry;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.ValueOperations;

import java.time.Instant;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class UrlServiceTest {

    @Mock UrlRepository repo;
    @Mock StringRedisTemplate redis;
    @Mock ValueOperations<String, String> valueOps;

    private UrlServiceImpl service;

    @BeforeEach
    void setUp() {
        AppProperties props = new AppProperties(
                "http://localhost", 7,
                new AppProperties.RateLimit(100)
        );
        UrlMetrics metrics = new UrlMetrics(new SimpleMeterRegistry());
        service = new UrlServiceImpl(repo, redis, new ShortCodeGenerator(), metrics, props);
        lenient().when(redis.opsForValue()).thenReturn(valueOps);
    }

    @Test
    void create_generatesCodeAndSaves() {
        when(repo.existsByShortCode(anyString())).thenReturn(false);
        Url saved = urlWith("abc1234", "https://example.com");
        when(repo.save(any())).thenReturn(saved);

        var resp = service.create(
                new CreateUrlRequest("https://example.com", null, null), "agent");

        assertThat(resp.shortCode()).hasSize(7);
        assertThat(resp.originalUrl()).isEqualTo("https://example.com");
        assertThat(resp.shortUrl()).startsWith("http://localhost/");
    }

    @Test
    void create_usesCustomCode() {
        when(repo.existsByShortCode("mycode")).thenReturn(false);
        when(repo.save(any())).thenReturn(urlWith("mycode", "https://example.com"));

        var resp = service.create(
                new CreateUrlRequest("https://example.com", "mycode", null), "");

        assertThat(resp.shortCode()).isEqualTo("mycode");
    }

    @Test
    void create_throwsWhenCustomCodeTaken() {
        when(repo.existsByShortCode("taken")).thenReturn(true);

        assertThatThrownBy(() ->
                service.create(new CreateUrlRequest("https://x.com", "taken", null), ""))
                .isInstanceOf(CodeAlreadyTakenException.class);
    }

    @Test
    void resolve_returnsCachedUrl() {
        when(valueOps.get("url:abc1234")).thenReturn("https://cached.com");

        assertThat(service.resolve("abc1234")).isEqualTo("https://cached.com");
        verify(repo, never()).findActive(any(), any());
    }

    @Test
    void resolve_fetchesFromDbOnCacheMiss() {
        when(valueOps.get("url:db1234x")).thenReturn(null);
        when(repo.findActive(eq("db1234x"), any()))
                .thenReturn(Optional.of(urlWith("db1234x", "https://fromdb.com")));

        assertThat(service.resolve("db1234x")).isEqualTo("https://fromdb.com");
    }

    @Test
    void resolve_throwsWhenNotFound() {
        when(valueOps.get(any())).thenReturn(null);
        when(repo.findActive(any(), any())).thenReturn(Optional.empty());

        assertThatThrownBy(() -> service.resolve("missing"))
                .isInstanceOf(UrlNotFoundException.class);
    }

    @Test
    void stats_returnsClickCount() {
        Url url = urlWith("xyz", "https://x.com");
        when(repo.findByShortCode("xyz")).thenReturn(Optional.of(url));

        var stats = service.stats("xyz");
        assertThat(stats.shortCode()).isEqualTo("xyz");
        assertThat(stats.clickCount()).isZero();
    }

    @Test
    void stats_throwsWhenNotFound() {
        when(repo.findByShortCode("nope")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> service.stats("nope"))
                .isInstanceOf(UrlNotFoundException.class);
    }

    private static Url urlWith(String code, String originalUrl) {
        Url url = new Url();
        url.setShortCode(code);
        url.setOriginalUrl(originalUrl);
        // simulate @PrePersist
        try {
            var f = Url.class.getDeclaredField("createdAt");
            f.setAccessible(true);
            f.set(url, Instant.now());
        } catch (Exception ignored) {}
        return url;
    }
}
