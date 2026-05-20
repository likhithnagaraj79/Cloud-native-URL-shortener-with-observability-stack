package com.urlshortener.worker;

import com.urlshortener.repository.UrlRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;

@Component
public class CleanupWorker {

    private static final Logger log = LoggerFactory.getLogger(CleanupWorker.class);
    private final UrlRepository repo;

    public CleanupWorker(UrlRepository repo) {
        this.repo = repo;
    }

    @Scheduled(fixedDelayString = "${app.cleanup-interval-ms:3600000}")
    @Transactional
    public void sweep() {
        int deleted = repo.deleteExpired(Instant.now());
        if (deleted > 0) {
            log.info("Cleanup sweep: deleted {} expired URLs", deleted);
        }
    }
}
