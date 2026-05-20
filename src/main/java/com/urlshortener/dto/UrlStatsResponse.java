package com.urlshortener.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.time.Instant;

public record UrlStatsResponse(
        @JsonProperty("short_code")   String shortCode,
        @JsonProperty("original_url") String originalUrl,
        @JsonProperty("click_count")  long clickCount,
        @JsonProperty("created_at")   Instant createdAt,
        @JsonProperty("expires_at")   Instant expiresAt
) {}
