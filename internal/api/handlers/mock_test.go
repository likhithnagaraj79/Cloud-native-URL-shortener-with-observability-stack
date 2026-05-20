package handlers_test

import (
	"context"
	"errors"

	"github.com/likhi/url-shortener/internal/models"
	"github.com/likhi/url-shortener/internal/repository"
	"github.com/likhi/url-shortener/internal/service"
)

type mockSvc struct {
	createResp *models.CreateURLResponse
	createErr  error
	resolveURL string
	resolveErr error
	statsResp  *models.URLStats
	statsErr   error
}

func (m *mockSvc) CreateShortURL(_ context.Context, _ *models.CreateURLRequest, _ string) (*models.CreateURLResponse, error) {
	return m.createResp, m.createErr
}

func (m *mockSvc) Resolve(_ context.Context, _ string) (string, error) {
	return m.resolveURL, m.resolveErr
}

func (m *mockSvc) GetStats(_ context.Context, _ string) (*models.URLStats, error) {
	return m.statsResp, m.statsErr
}

var _ service.URLShortener = (*mockSvc)(nil)

var errNotFound = repository.ErrNotFound
var errCodeTaken = service.ErrCodeTaken
var errInternal = errors.New("db exploded")
