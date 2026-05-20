package com.urlshortener.repository;

import com.urlshortener.model.Url;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.Optional;

@Repository
public interface UrlRepository extends JpaRepository<Url, Long> {

    @Query("SELECT u FROM Url u WHERE u.shortCode = :code AND (u.expiresAt IS NULL OR u.expiresAt > :now)")
    Optional<Url> findActive(@Param("code") String code, @Param("now") Instant now);

    Optional<Url> findByShortCode(String shortCode);

    boolean existsByShortCode(String shortCode);

    @Modifying
    @Query("UPDATE Url u SET u.clickCount = u.clickCount + 1 WHERE u.shortCode = :code")
    void incrementClickCount(@Param("code") String code);

    @Modifying
    @Query("DELETE FROM Url u WHERE u.expiresAt IS NOT NULL AND u.expiresAt < :now")
    int deleteExpired(@Param("now") Instant now);
}
