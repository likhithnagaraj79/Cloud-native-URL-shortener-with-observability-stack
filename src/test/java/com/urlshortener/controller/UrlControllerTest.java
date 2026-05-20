package com.urlshortener.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.urlshortener.dto.CreateUrlRequest;
import com.urlshortener.dto.CreateUrlResponse;
import com.urlshortener.dto.UrlStatsResponse;
import com.urlshortener.exception.CodeAlreadyTakenException;
import com.urlshortener.exception.GlobalExceptionHandler;
import com.urlshortener.exception.UrlNotFoundException;
import com.urlshortener.service.UrlService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.MediaType;
import org.springframework.http.converter.json.MappingJackson2HttpMessageConverter;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.validation.beanvalidation.LocalValidatorFactoryBean;

import java.time.Instant;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@ExtendWith(MockitoExtension.class)
class UrlControllerTest {

    @Mock UrlService service;

    private MockMvc mvc;

    private final ObjectMapper mapper = new ObjectMapper()
            .registerModule(new JavaTimeModule())
            .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);

    @BeforeEach
    void setUp() {
        LocalValidatorFactoryBean validator = new LocalValidatorFactoryBean();
        validator.afterPropertiesSet();

        mvc = MockMvcBuilders
                .standaloneSetup(new UrlController(service))
                .setControllerAdvice(new GlobalExceptionHandler())
                .setValidator(validator)
                .setMessageConverters(new MappingJackson2HttpMessageConverter(mapper))
                .build();
    }

    // ── POST /api/v1/urls ──────────────────────────────────────────────────

    @Test
    void create_returns201() throws Exception {
        when(service.create(any(), any())).thenReturn(new CreateUrlResponse(
                "abc1234", "http://localhost/abc1234",
                "https://example.com", Instant.now(), null));

        mvc.perform(post("/api/v1/urls")
                .contentType(MediaType.APPLICATION_JSON)
                .content(mapper.writeValueAsString(
                        new CreateUrlRequest("https://example.com", null, null))))
                .andExpect(status().isCreated())
                .andExpect(jsonPath("$.short_code").value("abc1234"));
    }

    @Test
    void create_returns400ForInvalidUrl() throws Exception {
        mvc.perform(post("/api/v1/urls")
                .contentType(MediaType.APPLICATION_JSON)
                .content(mapper.writeValueAsString(
                        new CreateUrlRequest("not-a-url", null, null))))
                .andExpect(status().isBadRequest());
    }

    @Test
    void create_returns409WhenCodeTaken() throws Exception {
        when(service.create(any(), any())).thenThrow(new CodeAlreadyTakenException("taken"));

        mvc.perform(post("/api/v1/urls")
                .contentType(MediaType.APPLICATION_JSON)
                .content(mapper.writeValueAsString(
                        new CreateUrlRequest("https://example.com", "taken", null))))
                .andExpect(status().isConflict());
    }

    // ── GET /:code ─────────────────────────────────────────────────────────

    @Test
    void redirect_returns301() throws Exception {
        when(service.resolve("abc1234")).thenReturn("https://destination.com");

        mvc.perform(get("/abc1234"))
                .andExpect(status().isMovedPermanently())
                .andExpect(header().string("Location", "https://destination.com"));
    }

    @Test
    void redirect_returns404WhenNotFound() throws Exception {
        when(service.resolve("missing")).thenThrow(new UrlNotFoundException("missing"));

        mvc.perform(get("/missing"))
                .andExpect(status().isNotFound());
    }

    // ── GET /api/v1/urls/:code/stats ───────────────────────────────────────

    @Test
    void stats_returns200WithClickCount() throws Exception {
        when(service.stats("abc1234")).thenReturn(new UrlStatsResponse(
                "abc1234", "https://example.com", 42, Instant.now(), null));

        mvc.perform(get("/api/v1/urls/abc1234/stats"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.click_count").value(42));
    }

    @Test
    void stats_returns404WhenNotFound() throws Exception {
        when(service.stats("nope")).thenThrow(new UrlNotFoundException("nope"));

        mvc.perform(get("/api/v1/urls/nope/stats"))
                .andExpect(status().isNotFound());
    }

    // ── GET /health ────────────────────────────────────────────────────────

    @Test
    void health_returns200() throws Exception {
        mvc.perform(get("/health"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.status").value("ok"));
    }
}
