package com.urlshortener.service;

import com.urlshortener.dto.CreateUrlRequest;
import com.urlshortener.dto.CreateUrlResponse;
import com.urlshortener.dto.UrlStatsResponse;

public interface UrlService {
    CreateUrlResponse create(CreateUrlRequest request, String userAgent);
    String resolve(String shortCode);
    UrlStatsResponse stats(String shortCode);
}
