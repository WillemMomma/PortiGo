package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go-gateway/internal/config"
	httpx "go-gateway/internal/http"
	"go-gateway/internal/services/models"
	"go-gateway/internal/services/proxy"
	"go-gateway/internal/store/postgres"
)

func main() {
    slog.Info("Starting API server")

    cfg := config.Load()

    db, err := postgres.Open(cfg.DatabaseURL)
    if err != nil {
        slog.Error("db open failed", slog.Any("err", err))
        os.Exit(1)
    }

    modelRepo := postgres.NewModelRepository(db)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := modelRepo.EnsureSchema(ctx); err != nil {
        slog.Error("ensure schema failed", slog.Any("err", err))
        os.Exit(1)
    }

    modelSvc := models.NewService(modelRepo)
    proxySvc := proxy.NewService(modelRepo)

    handler := httpx.Handler{Models: modelSvc, Proxy: proxySvc}
    router := httpx.NewRouter(handler)

    srv := &http.Server{
        Addr:              cfg.Addr,
        Handler:           router,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       15 * time.Second,
        WriteTimeout:      30 * time.Second,
        IdleTimeout:       60 * time.Second,
    }

    slog.Info("Listening", slog.String("addr", cfg.Addr))
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        slog.Error("server error", slog.Any("err", err))
        os.Exit(1)
    }
}