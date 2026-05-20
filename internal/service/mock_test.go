package service_test

import (
	"context"
	"errors"
	"time"

	"github.com/likhi/url-shortener/internal/models"
	"github.com/likhi/url-shortener/internal/repository"
)

// --- mock URLStore ---

type mockStore struct {
	urls   map[string]*models.URL
	create func(*models.URL) error
}

func newMockStore() *mockStore {
	return &mockStore{urls: make(map[string]*models.URL)}
}

func (m *mockStore) Create(ctx context.Context, url *models.URL) error {
	if m.create != nil {
		return m.create(url)
	}
	url.ID = 1
	url.CreatedAt = time.Now()
	url.ClickCount = 0
	m.urls[url.ShortCode] = url
	return nil
}

func (m *mockStore) GetByShortCode(ctx context.Context, code string) (*models.URL, error) {
	if url, ok := m.urls[code]; ok {
		return url, nil
	}
	return nil, repository.ErrNotFound
}

func (m *mockStore) IncrementClickCount(ctx context.Context, code string) error {
	if url, ok := m.urls[code]; ok {
		url.ClickCount++
	}
	return nil
}

func (m *mockStore) ShortCodeExists(ctx context.Context, code string) (bool, error) {
	_, ok := m.urls[code]
	return ok, nil
}

func (m *mockStore) GetStats(ctx context.Context, code string) (*models.URLStats, error) {
	if url, ok := m.urls[code]; ok {
		return &models.URLStats{
			ShortCode:   url.ShortCode,
			OriginalURL: url.OriginalURL,
			ClickCount:  url.ClickCount,
			CreatedAt:   url.CreatedAt,
			ExpiresAt:   url.ExpiresAt,
		}, nil
	}
	return nil, repository.ErrNotFound
}

func (m *mockStore) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

// --- mock Cache ---

type mockCache struct {
	data map[string]string
	err  error
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]string)}
}

func (c *mockCache) Get(ctx context.Context, key string) (string, error) {
	if c.err != nil {
		return "", c.err
	}
	val, ok := c.data[key]
	if !ok {
		return "", errors.New("cache miss")
	}
	return val, nil
}

func (c *mockCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	c.data[key] = value
	return nil
}
