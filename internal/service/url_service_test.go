package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/likhi/url-shortener/internal/models"
	"github.com/likhi/url-shortener/internal/repository"
	"github.com/likhi/url-shortener/internal/service"
)

func newService(store *mockStore, cache *mockCache) *service.URLService {
	return service.NewURLService(store, cache, zap.NewNop(), "http://localhost", 7)
}

func TestCreateShortURL_GeneratesCode(t *testing.T) {
	svc := newService(newMockStore(), newMockCache())

	resp, err := svc.CreateShortURL(context.Background(), &models.CreateURLRequest{
		OriginalURL: "https://example.com",
	}, "test-agent")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.ShortCode) != 7 {
		t.Errorf("short code length = %d, want 7", len(resp.ShortCode))
	}
	if resp.OriginalURL != "https://example.com" {
		t.Errorf("original url = %q", resp.OriginalURL)
	}
	if !strings.HasPrefix(resp.ShortURL, "http://localhost/") {
		t.Errorf("short url = %q, want prefix http://localhost/", resp.ShortURL)
	}
}

func TestCreateShortURL_CustomCode(t *testing.T) {
	svc := newService(newMockStore(), newMockCache())

	resp, err := svc.CreateShortURL(context.Background(), &models.CreateURLRequest{
		OriginalURL: "https://example.com",
		CustomCode:  "mycode",
	}, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ShortCode != "mycode" {
		t.Errorf("short code = %q, want mycode", resp.ShortCode)
	}
}

func TestCreateShortURL_CustomCodeConflict(t *testing.T) {
	store := newMockStore()
	svc := newService(store, newMockCache())

	req := &models.CreateURLRequest{OriginalURL: "https://example.com", CustomCode: "taken"}
	svc.CreateShortURL(context.Background(), req, "")

	_, err := svc.CreateShortURL(context.Background(), req, "")
	if !errors.Is(err, service.ErrCodeTaken) {
		t.Errorf("expected ErrCodeTaken, got %v", err)
	}
}

func TestCreateShortURL_DBError(t *testing.T) {
	store := newMockStore()
	store.create = func(*models.URL) error { return errors.New("db error") }

	svc := newService(store, newMockCache())
	_, err := svc.CreateShortURL(context.Background(), &models.CreateURLRequest{
		OriginalURL: "https://example.com",
	}, "")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolve_CacheHit(t *testing.T) {
	cache := newMockCache()
	cache.data["url:abc1234"] = "https://cached.com"

	svc := newService(newMockStore(), cache)
	got, err := svc.Resolve(context.Background(), "abc1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://cached.com" {
		t.Errorf("got %q, want https://cached.com", got)
	}
}

func TestResolve_CacheMissFallsBackToDB(t *testing.T) {
	store := newMockStore()
	store.urls["db1234x"] = &models.URL{
		ShortCode:   "db1234x",
		OriginalURL: "https://fromdb.com",
	}

	svc := newService(store, newMockCache())
	got, err := svc.Resolve(context.Background(), "db1234x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://fromdb.com" {
		t.Errorf("got %q, want https://fromdb.com", got)
	}
}

func TestResolve_NotFound(t *testing.T) {
	svc := newService(newMockStore(), newMockCache())
	_, err := svc.Resolve(context.Background(), "missing")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetStats_Found(t *testing.T) {
	store := newMockStore()
	store.urls["xyz"] = &models.URL{ShortCode: "xyz", OriginalURL: "https://x.com", ClickCount: 5}

	svc := newService(store, newMockCache())
	stats, err := svc.GetStats(context.Background(), "xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.ClickCount != 5 {
		t.Errorf("click count = %d, want 5", stats.ClickCount)
	}
}

func TestGetStats_NotFound(t *testing.T) {
	svc := newService(newMockStore(), newMockCache())
	_, err := svc.GetStats(context.Background(), "nope")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
