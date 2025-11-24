package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"webhook-receiver/internal/client"
	"webhook-receiver/internal/config"
	"webhook-receiver/internal/model"
)

type BatchProcessor struct {
	cfg         *config.Config
	client      *client.HttpClient
	log         *zap.Logger
	ticker      *time.Ticker
	buffer      []model.LogEntry
	mu          sync.Mutex
	flushSignal chan struct{}

	ErrChan chan error
}

func NewBatchProcessor(cfg *config.Config, client *client.HttpClient, log *zap.Logger) *BatchProcessor {
	return &BatchProcessor{
		cfg:         cfg,
		client:      client,
		log:         log,
		buffer:      make([]model.LogEntry, 0, cfg.BatchSize),
		flushSignal: make(chan struct{}, 1),
		ErrChan:     make(chan error, 1),
	}
}

func (bp *BatchProcessor) Start(ctx context.Context) {
	go func() {
		bp.ticker = time.NewTicker(bp.cfg.BatchInterval)
		defer bp.ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-bp.ticker.C:
				bp.log.Info("Flushing buffer by ticker")
				bp.flush(ctx)
			case <-bp.flushSignal:
				bp.log.Info("Flushing buffer by batch")
				bp.flush(ctx)
				bp.ticker.Reset(bp.cfg.BatchInterval)
			}
		}
	}()
}

func (bp *BatchProcessor) Add(entry model.LogEntry) {
	bp.mu.Lock()

	bp.buffer = append(bp.buffer, entry)
	shouldFlush := len(bp.buffer) >= bp.cfg.BatchSize

	bp.mu.Unlock()

	if shouldFlush {
		select {
		case bp.flushSignal <- struct{}{}:
		default:
		}
	}
}

func (bp *BatchProcessor) flush(ctx context.Context) {
	bp.mu.Lock()

	if len(bp.buffer) == 0 {
		bp.log.Info("Buffer is empty, skipping flush")
		bp.mu.Unlock()
		return
	}

	buffer := bp.buffer
	bp.buffer = make([]model.LogEntry, 0, bp.cfg.BatchSize)
	bp.mu.Unlock()

	if err := bp.sendBatch(ctx, buffer); err != nil {
		select {
		case bp.ErrChan <- err:
		default:
			bp.log.Warn("Error channel is full")
		}
	}
}

func (bp *BatchProcessor) sendBatch(ctx context.Context, batch []model.LogEntry) error {
	bp.log.Info("Processing buffer", zap.Int("size", len(batch)))

	payload, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := bp.client.PostWithRetry(ctx, payload); err != nil {
		return fmt.Errorf("client.PostWithRetry: %w", err)
	}

	return nil
}
