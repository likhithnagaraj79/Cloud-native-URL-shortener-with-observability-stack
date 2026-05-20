package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/likhi/url-shortener/internal/api"
	"github.com/likhi/url-shortener/internal/config"
	"github.com/likhi/url-shortener/internal/database"
	"github.com/likhi/url-shortener/internal/repository"
	"github.com/likhi/url-shortener/internal/service"
	"github.com/likhi/url-shortener/internal/worker"
	"github.com/likhi/url-shortener/migrations"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	db, err := database.NewPostgresPool(&cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer db.Close()

	ctx := context.Background()

	logger.Info("running database migrations")
	if err := database.RunMigrations(ctx, db, migrations.FS); err != nil {
		logger.Fatal("migrations failed", zap.Error(err))
	}
	logger.Info("migrations up to date")

	redisClient, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		logger.Fatal("failed to connect to redis", zap.Error(err))
	}
	defer redisClient.Close()

	cache := database.NewRedisCache(redisClient)
	repo := repository.NewURLRepository(db)
	svc := service.NewURLService(repo, cache, logger, cfg.App.BaseURL, cfg.App.ShortCodeLen)
	router := api.NewRouter(svc, redisClient, logger)

	cleanupCtx, cancelCleanup := context.WithCancel(ctx)
	defer cancelCleanup()
	go worker.NewCleanupWorker(repo, 1*time.Hour, logger).Run(cleanupCtx)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")
	cancelCleanup()

	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}
	logger.Info("server exited")
}
