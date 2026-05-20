package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/likhi/url-shortener/pkg/metrics"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		metrics.ActiveConnections.Inc()

		c.Next()

		metrics.ActiveConnections.Dec()
		metrics.HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			strconv.Itoa(c.Writer.Status()),
		).Observe(time.Since(start).Seconds())
	}
}
