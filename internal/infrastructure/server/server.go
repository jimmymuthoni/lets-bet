package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

// RunHTTP starts an HTTP server on the given address with the given handler
// and blocks until the context is cancelled or an OS signal (SIGINT/SIGTERM)
// is received, then performs a graceful shutdown with a 15s drain window.
func RunHTTP(ctx context.Context, addr string, handler http.Handler, logger *slog.Logger) error {
	// Derive a context that is cancelled on SIGINT/SIGTERM. This is the idiomatic
	// Go 1.16+ replacement for manual signal.Notify + channel plumbing.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	// Buffered so the goroutine never blocks on send if the caller returns early.
	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown requested", "cause", context.Cause(ctx))
	case err, ok := <-errCh:
		if ok && err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("server shut down cleanly")
	return nil
}
