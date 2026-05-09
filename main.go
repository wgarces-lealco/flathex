package main

import (
	"context"
	"errors"
	"fmt"
	"flathex/internal/bootstrap"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := bootstrap.LoadConfig()
	bootstrap.InitLogger(cfg)

	// ── 1. Context raíz ligado a señales del OS ──────────────────────────────
	// Cuando el orquestador (Kubernetes, Docker) envía SIGTERM o el dev
	// presiona Ctrl+C, ctx se cancela y el servidor drena las requests en vuelo.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := bootstrap.OpenDB(cfg)
	if err != nil {
		return fmt.Errorf("open database %q: %w", cfg.SQLitePath, err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			slog.Error("database close", "error", cerr)
		}
	}()

	e := bootstrap.BuildEcho(cfg, db)

	slog.Info("TaskHex starting", "port", cfg.Port, "env", cfg.Environment, "db", cfg.SQLitePath)
	bootstrap.PrintRoutes(e)

	// ── 1. Arranque + graceful shutdown ──────────────────────────────────────
	// Start corre en su propia goroutine; esperamos ctx.Done() (señal OS).
	// Shutdown da a Echo tiempo de terminar requests en vuelo antes de salir.
	errCh := make(chan error, 1)
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http server: %w", err)
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err // el servidor falló antes de recibir señal
	case <-ctx.Done():
		slog.Info("shutdown signal received, draining requests...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	slog.Info("shutdown complete")
	return <-errCh // nil si Start se cerró limpiamente
}
