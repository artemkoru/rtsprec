package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artemkoru/rtsprec/internal/app"
)

const shutdownTimeout = 5 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("starting RTSP surveillance system")

	// Context is cancelled on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	if err := application.Init(ctx); err != nil {
		logger.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}

	// Shut down when a signal arrives; Run returns once the server has stopped
	go func() {
		<-ctx.Done()
		logger.Info("shutting down gracefully")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := application.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown error", "error", err)
		}
	}()

	if err := application.Run(); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}

	logger.Info("shutdown complete")
}
