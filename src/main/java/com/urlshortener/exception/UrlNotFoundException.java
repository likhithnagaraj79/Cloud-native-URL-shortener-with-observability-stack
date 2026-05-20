package com.urlshortener.exception;

public class UrlNotFoundException extends RuntimeException {
    public UrlNotFoundException(String code) {
        super("Short URL not found or expired: " + code);
    }
}
