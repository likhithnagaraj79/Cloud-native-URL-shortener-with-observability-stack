package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/likhi/url-shortener/internal/models"
)

var ErrNotFound = errors.New("url not found")

type URLRepository struct {
	db *pgxpool.Pool
}

func NewURLRepository(db *pgxpool.Pool) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Create(ctx context.Context, url *models.URL) error {
	query := `
		INSERT INTO urls (short_code, original_url, expires_at, user_agent)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, click_count`

	return r.db.QueryRow(ctx, query,
		url.ShortCode, url.OriginalURL, url.ExpiresAt, url.UserAgent,
	).Scan(&url.ID, &url.CreatedAt, &url.ClickCount)
}

func (r *URLRepository) GetByShortCode(ctx context.Context, code string) (*models.URL, error) {
	url := &models.URL{}
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, click_count
		FROM urls
		WHERE short_code = $1 AND (expires_at IS NULL OR expires_at > NOW())`

	err := r.db.QueryRow(ctx, query, code).Scan(
		&url.ID, &url.ShortCode, &url.OriginalURL,
		&url.CreatedAt, &url.ExpiresAt, &url.ClickCount,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return url, nil
}

func (r *URLRepository) IncrementClickCount(ctx context.Context, code string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE urls SET click_count = click_count + 1 WHERE short_code = $1`, code,
	)
	return err
}

func (r *URLRepository) ShortCodeExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`, code,
	).Scan(&exists)
	return exists, err
}

func (r *URLRepository) GetStats(ctx context.Context, code string) (*models.URLStats, error) {
	stats := &models.URLStats{}
	query := `
		SELECT short_code, original_url, click_count, created_at, expires_at
		FROM urls WHERE short_code = $1`

	err := r.db.QueryRow(ctx, query, code).Scan(
		&stats.ShortCode, &stats.OriginalURL,
		&stats.ClickCount, &stats.CreatedAt, &stats.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return stats, nil
}

func (r *URLRepository) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM urls WHERE expires_at IS NOT NULL AND expires_at < $1`, time.Now(),
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
