package com.urlshortener.exception;

public class CodeAlreadyTakenException extends RuntimeException {
    public CodeAlreadyTakenException(String code) {
        super("Custom code already taken: " + code);
    }
}
