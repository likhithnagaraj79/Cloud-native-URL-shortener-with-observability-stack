package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/likhi/url-shortener/internal/api/handlers"
	"github.com/likhi/url-shortener/internal/api/middleware"
	"github.com/likhi/url-shortener/internal/service"
)

func NewRouter(svc service.URLShortener, rdb *redis.Client, logger *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(middleware.CORS())
	r.Use(middleware.ZapLogger(logger))
	r.Use(middleware.PrometheusMiddleware())
	r.Use(gin.Recovery())

	h := handlers.NewURLHandler(svc, logger)

	r.GET("/health", h.Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 100 requests per minute per IP across all replicas
	rl := middleware.RateLimiter(rdb, 100, time.Minute)

	v1 := r.Group("/api/v1")
	v1.Use(rl)
	{
		v1.POST("/urls", h.Create)
		v1.GET("/urls/:code/stats", h.Stats)
	}

	r.GET("/:code", rl, h.Redirect)

	return r
}
