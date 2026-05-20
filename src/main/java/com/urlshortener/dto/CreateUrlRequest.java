package com.urlshortener.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import jakarta.validation.constraints.NotBlank;
import org.hibernate.validator.constraints.URL;

import java.time.Instant;

public record CreateUrlRequest(
        @JsonProperty("original_url")
        @NotBlank(message = "original_url is required")
        @URL(message = "original_url must be a valid URL")
        String originalUrl,

        @JsonProperty("custom_code")
        String customCode,

        @JsonProperty("expires_at")
        Instant expiresAt
) {}
