package com.urlshortener.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.time.Instant;

public record CreateUrlResponse(
        @JsonProperty("short_code")  String shortCode,
        @JsonProperty("short_url")   String shortUrl,
        @JsonProperty("original_url") String originalUrl,
        @JsonProperty("created_at")  Instant createdAt,
        @JsonProperty("expires_at")  Instant expiresAt
) {}
