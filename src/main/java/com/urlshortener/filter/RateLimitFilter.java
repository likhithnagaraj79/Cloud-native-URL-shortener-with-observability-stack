package com.urlshortener.filter;

import com.urlshortener.config.AppProperties;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.core.annotation.Order;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.time.Duration;

@Component
@Order(1)
public class RateLimitFilter extends OncePerRequestFilter {

    private static final Duration WINDOW = Duration.ofMinutes(1);

    private final StringRedisTemplate redis;
    private final int limit;

    public RateLimitFilter(StringRedisTemplate redis, AppProperties props) {
        this.redis = redis;
        this.limit = props.rateLimit().requestsPerMinute();
    }

    @Override
    protected void doFilterInternal(HttpServletRequest req,
                                    HttpServletResponse res,
                                    FilterChain chain) throws ServletException, IOException {
        String key = "rl:" + req.getRemoteAddr();
        Long count;
        try {
            count = redis.opsForValue().increment(key);
            if (count != null && count == 1) {
                redis.expire(key, WINDOW);
            }
        } catch (Exception e) {
            // fail open — don't block traffic if Redis is unavailable
            chain.doFilter(req, res);
            return;
        }

        int remaining = (count == null) ? limit : Math.max(0, limit - count.intValue());
        res.setHeader("X-RateLimit-Limit", String.valueOf(limit));
        res.setHeader("X-RateLimit-Remaining", String.valueOf(remaining));

        if (count != null && count > limit) {
            res.setStatus(429);
            res.setContentType(MediaType.APPLICATION_JSON_VALUE);
            res.getWriter().write("{\"error\":\"rate limit exceeded\"}");
            return;
        }
        chain.doFilter(req, res);
    }
}
