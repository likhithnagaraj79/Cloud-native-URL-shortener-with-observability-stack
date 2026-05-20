package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/likhi/url-shortener/internal/repository"
)

// CleanupWorker deletes expired URLs on a fixed interval.
type CleanupWorker struct {
	repo     repository.URLStore
	interval time.Duration
	logger   *zap.Logger
}

func NewCleanupWorker(repo repository.URLStore, interval time.Duration, logger *zap.Logger) *CleanupWorker {
	return &CleanupWorker{repo: repo, interval: interval, logger: logger}
}

// Run blocks until ctx is cancelled. Call it in a goroutine.
func (w *CleanupWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("cleanup worker started", zap.Duration("interval", w.interval))

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("cleanup worker stopped")
			return
		case <-ticker.C:
			w.sweep(ctx)
		}
	}
}

func (w *CleanupWorker) sweep(ctx context.Context) {
	n, err := w.repo.DeleteExpired(ctx)
	if err != nil {
		w.logger.Error("cleanup sweep failed", zap.Error(err))
		return
	}
	if n > 0 {
		w.logger.Info("cleanup sweep", zap.Int64("deleted", n))
	}
}
