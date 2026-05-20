package com.urlshortener.controller;

import com.urlshortener.dto.CreateUrlRequest;
import com.urlshortener.dto.CreateUrlResponse;
import com.urlshortener.dto.UrlStatsResponse;
import com.urlshortener.service.UrlService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.net.URI;
import java.util.Map;

@RestController
public class UrlController {

    private final UrlService service;

    public UrlController(UrlService service) {
        this.service = service;
    }

    @PostMapping("/api/v1/urls")
    public ResponseEntity<CreateUrlResponse> create(
            @Valid @RequestBody CreateUrlRequest request,
            @RequestHeader(value = "User-Agent", defaultValue = "") String userAgent) {

        CreateUrlResponse response = service.create(request, userAgent);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @GetMapping("/{code}")
    public ResponseEntity<Void> redirect(@PathVariable String code) {
        String originalUrl = service.resolve(code);
        return ResponseEntity.status(HttpStatus.MOVED_PERMANENTLY)
                .location(URI.create(originalUrl))
                .build();
    }

    @GetMapping("/api/v1/urls/{code}/stats")
    public ResponseEntity<UrlStatsResponse> stats(@PathVariable String code) {
        return ResponseEntity.ok(service.stats(code));
    }

    @GetMapping("/health")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "ok"));
    }
}
