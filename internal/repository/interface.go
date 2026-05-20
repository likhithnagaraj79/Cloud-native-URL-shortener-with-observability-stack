package repository

import (
	"context"

	"github.com/likhi/url-shortener/internal/models"
)

type URLStore interface {
	Create(ctx context.Context, url *models.URL) error
	GetByShortCode(ctx context.Context, code string) (*models.URL, error)
	IncrementClickCount(ctx context.Context, code string) error
	ShortCodeExists(ctx context.Context, code string) (bool, error)
	GetStats(ctx context.Context, code string) (*models.URLStats, error)
	DeleteExpired(ctx context.Context) (int64, error)
}
