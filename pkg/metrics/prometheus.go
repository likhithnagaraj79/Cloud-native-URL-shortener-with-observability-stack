package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	URLsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "url_shortener_urls_created_total",
		Help: "Total number of short URLs created",
	})

	URLRedirectsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "url_shortener_redirects_total",
		Help: "Total number of URL redirects by short code",
	}, []string{"short_code"})

	CacheHitsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "url_shortener_cache_hits_total",
		Help: "Total Redis cache hits",
	})

	CacheMissesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "url_shortener_cache_misses_total",
		Help: "Total Redis cache misses",
	})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "url_shortener_http_request_duration_seconds",
		Help:    "HTTP request latency",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	ActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "url_shortener_active_connections",
		Help: "Number of active HTTP connections",
	})
)
