package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"webhook-receiver/internal/config"
)

type MaxRetriesExceededError struct {
	Count int
}

func (e *MaxRetriesExceededError) Error() string {
	return fmt.Sprintf("failed to send request after %d retries", e.Count)
}

type HttpClient struct {
	client *http.Client
	cfg    *config.Config
	log    *zap.Logger
}

func NewHttpClient(cfg *config.Config, log *zap.Logger) *HttpClient {
	return &HttpClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 30 * time.Second,
			},
		},
		cfg: cfg,
		log: log,
	}
}

func (c *HttpClient) PostWithRetry(ctx context.Context, payload []byte) error {
	start := time.Now()

	for i := 1; i <= c.cfg.RetryAttempts; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.PostEndpoint, bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			c.log.Warn("Failed to send request", zap.Error(err), zap.Int("attempt", i))
			time.Sleep(c.cfg.RetryDelay)

			continue
		}

		if _, err = io.Copy(io.Discard, resp.Body); err != nil {
			c.log.Warn("Failed to read response body", zap.Error(err))
		}
		if err = resp.Body.Close(); err != nil {
			c.log.Warn("Failed to close response body", zap.Error(err))
		}

		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			c.log.Info(
				"Request sent successfully",
				zap.Int("status_code", resp.StatusCode),
				zap.String("duration", time.Since(start).String()),
			)

			return nil
		}

		c.log.Warn(
			"Request failed to send",
			zap.Int("status_code", resp.StatusCode),
			zap.String("duration", time.Since(start).String()),
		)

		time.Sleep(c.cfg.RetryDelay)
	}

	return &MaxRetriesExceededError{
		Count: c.cfg.RetryAttempts,
	}
}
