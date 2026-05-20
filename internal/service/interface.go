package service

import (
	"context"

	"github.com/likhi/url-shortener/internal/models"
)

type URLShortener interface {
	CreateShortURL(ctx context.Context, req *models.CreateURLRequest, userAgent string) (*models.CreateURLResponse, error)
	Resolve(ctx context.Context, code string) (string, error)
	GetStats(ctx context.Context, code string) (*models.URLStats, error)
}
