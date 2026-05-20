package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/likhi/url-shortener/internal/database"
	"github.com/likhi/url-shortener/internal/models"
	"github.com/likhi/url-shortener/internal/repository"
	"github.com/likhi/url-shortener/pkg/metrics"
	"github.com/likhi/url-shortener/pkg/shortener"
)

var ErrCodeTaken = errors.New("custom code already taken")

const cacheTTL = 24 * time.Hour

type URLService struct {
	repo         repository.URLStore
	cache        database.Cache
	logger       *zap.Logger
	baseURL      string
	shortCodeLen int
}

func NewURLService(
	repo repository.URLStore,
	cache database.Cache,
	logger *zap.Logger,
	baseURL string,
	shortCodeLen int,
) *URLService {
	return &URLService{
		repo:         repo,
		cache:        cache,
		logger:       logger,
		baseURL:      baseURL,
		shortCodeLen: shortCodeLen,
	}
}

func (s *URLService) CreateShortURL(ctx context.Context, req *models.CreateURLRequest, userAgent string) (*models.CreateURLResponse, error) {
	code := req.CustomCode
	if code == "" {
		var err error
		code, err = s.generateUniqueCode(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}
	} else {
		exists, err := s.repo.ShortCodeExists(ctx, code)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCodeTaken
		}
	}

	url := &models.URL{
		ShortCode:   code,
		OriginalURL: req.OriginalURL,
		ExpiresAt:   req.ExpiresAt,
		UserAgent:   userAgent,
	}

	if err := s.repo.Create(ctx, url); err != nil {
		return nil, fmt.Errorf("failed to save url: %w", err)
	}

	s.cacheURL(ctx, url)
	metrics.URLsCreatedTotal.Inc()

	return &models.CreateURLResponse{
		ShortCode:   code,
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, code),
		OriginalURL: req.OriginalURL,
		CreatedAt:   url.CreatedAt,
		ExpiresAt:   url.ExpiresAt,
	}, nil
}

func (s *URLService) Resolve(ctx context.Context, code string) (string, error) {
	if cached, err := s.cache.Get(ctx, cacheKey(code)); err == nil {
		metrics.CacheHitsTotal.Inc()
		metrics.URLRedirectsTotal.WithLabelValues(code).Inc()
		go s.incrementClickAsync(code)
		return cached, nil
	}
	metrics.CacheMissesTotal.Inc()

	url, err := s.repo.GetByShortCode(ctx, code)
	if errors.Is(err, repository.ErrNotFound) {
		return "", repository.ErrNotFound
	}
	if err != nil {
		return "", err
	}

	s.cacheURL(ctx, url)
	metrics.URLRedirectsTotal.WithLabelValues(code).Inc()
	go s.incrementClickAsync(code)

	return url.OriginalURL, nil
}

func (s *URLService) GetStats(ctx context.Context, code string) (*models.URLStats, error) {
	return s.repo.GetStats(ctx, code)
}

func (s *URLService) generateUniqueCode(ctx context.Context) (string, error) {
	for range 5 {
		code, err := shortener.Generate(s.shortCodeLen)
		if err != nil {
			return "", err
		}
		exists, err := s.repo.ShortCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", errors.New("failed to generate unique code after 5 attempts")
}

func (s *URLService) cacheURL(ctx context.Context, url *models.URL) {
	ttl := cacheTTL
	if url.ExpiresAt != nil {
		remaining := time.Until(*url.ExpiresAt)
		if remaining <= 0 {
			return
		}
		if remaining < ttl {
			ttl = remaining
		}
	}
	s.cache.Set(ctx, cacheKey(url.ShortCode), url.OriginalURL, ttl)
}

func (s *URLService) incrementClickAsync(code string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.repo.IncrementClickCount(ctx, code); err != nil {
		s.logger.Warn("failed to increment click count", zap.String("code", code), zap.Error(err))
	}
}

func cacheKey(code string) string {
	return "url:" + code
}
