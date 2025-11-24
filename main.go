package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	httpClient "webhook-receiver/internal/client"
	"webhook-receiver/internal/config"
	"webhook-receiver/internal/handlers"
	"webhook-receiver/internal/processor"
	"webhook-receiver/internal/server"
	"webhook-receiver/pkg/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := logger.Init(); err != nil {
		panic(err)
	}
	defer logger.Sync()

	cfg := config.LoadConfig()
	logger.Log.Info(
		"Config loaded",
		zap.String("Port", cfg.Port),
		zap.String("Post Endpoint", cfg.PostEndpoint),
		zap.Int("Batch Size", cfg.BatchSize),
		zap.Duration("Batch Interval", cfg.BatchInterval),
		zap.Int("Retry Attempts", cfg.RetryAttempts),
		zap.Duration("Retry Delay", cfg.RetryDelay),
	)

	client := httpClient.NewHttpClient(cfg, logger.Log)

	bp := processor.NewBatchProcessor(cfg, client, logger.Log)
	bp.Start(ctx)

	logHandler := handlers.NewLogHandler(bp, logger.Log)

	srv := server.New(cfg, logHandler)

	go func() {
		logger.Log.Info("Starting server", zap.String("Addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Log.Info("Got signal", zap.String("Signal", sig.String()))
	case err := <-bp.ErrChan:
		logger.Log.Error("Error from batch processor", zap.Error(err))
	}

	logger.Log.Info("Shutting down server...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Fatal("Failed to shutdown server gracefully", zap.Error(err))
	}

	logger.Log.Info("Server stopped")
}
