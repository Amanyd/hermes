package engine

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/store"
)

type Job struct {
	RelayID string
	EventID string
	Payload []byte
	MsgAck  func(bool)
}

type WorkerPool struct {
	JobQueue   chan Job
	MaxWorkers int
	Store      *store.Store
	Registry   *Registry
	Logger     *slog.Logger
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// Constructor with dependency injxtn
func NewWorkerPool(maxWorkers int, db *store.Store, reg *Registry, logger *slog.Logger) *WorkerPool {
	return &WorkerPool{
		JobQueue:   make(chan Job, 100),
		MaxWorkers: maxWorkers,
		Store:      db,
		Registry:   reg,
		Logger:     logger,
	}
}

// Spaws all worker goroutines
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.ctx, wp.cancel = context.WithCancel(ctx)
	wp.Logger.Info("starting worker pool",
		slog.Int("max_workers", wp.MaxWorkers),
		slog.Int("queue_size", cap(wp.JobQueue)),
	)
	for i := 0; i < wp.MaxWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
	wp.Logger.Info("worker pool started",
		slog.Int("workers", wp.MaxWorkers))
}

// Each worker runs its own goroutine
func (wp *WorkerPool) worker(_ context.Context, id int) {
	defer wp.wg.Done()
	workerLogger := wp.Logger.With(slog.Int("worker_id", id))
	workerLogger.Debug("worker started")
	for {
		select {
		case <-wp.ctx.Done():
			workerLogger.Info("worker shutting down")
			return
		case job, ok := <-wp.JobQueue:
			if !ok {
				workerLogger.Info("job queue closed, worker exiting")
				return
			}
			start := time.Now()
			workerLogger.Info("processing relay",
				slog.String("relay_id", job.RelayID),
				slog.String("event_id", job.EventID))
			err := wp.process(wp.ctx, job, workerLogger)
			duration := time.Since(start)
			if err != nil {
				workerLogger.Error("relay execution failed",
					slog.String("relay_id", job.RelayID),
					slog.String("event_id", job.EventID),
					slog.Duration("duration", duration),
					slog.String("error", err.Error()))
				job.MsgAck(false)
			} else {
				workerLogger.Info("relay execution succeeded",
					slog.String("relay_id", job.RelayID),
					slog.String("event_id", job.EventID),
					slog.Duration("duration", duration))
				job.MsgAck(true)
			}
		}
	}
}

// Executes the actual workflow logic
func (wp *WorkerPool) process(ctx context.Context, job Job, logger *slog.Logger) (err error) {
	status := "success"
	details := "Relay executed successfully"

	if job.EventID != "" {
		isNew, dedupeErr := wp.Store.RegisterEvent(ctx, job.RelayID, job.EventID)
		if dedupeErr != nil {
			return dedupeErr
		}
		if !isNew {
			logger.Info("duplicate event skipped",
				slog.String("relay_id", job.RelayID),
				slog.String("event_id", job.EventID))
			return nil
		}
	}
	defer func() {
		logCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err != nil {
			status = "failed"
			details = err.Error()
		}
		logErr := wp.Store.LogExecution(logCtx, job.RelayID, status, details, job.EventID, job.Payload)
		if logErr != nil {
			logger.Error("failed to save execution log", slog.String("error", logErr.Error()))
		}
	}()
	userID, ownerErr := wp.Store.GetRelayOwner(ctx, job.RelayID)
	if ownerErr != nil {
		return ownerErr
	}

	actions, fetchErr := wp.Store.GetRelayActions(ctx, job.RelayID)
	if fetchErr != nil {
		return fetchErr
	}
	for _, act := range actions {
		resolved, resolveErr := wp.resolveSecrets(ctx, userID, act.Config)
		if resolveErr != nil {
			return fmt.Errorf("action %s (order %d) secret resolution failed: %w",
				act.ActionType, act.OrderIndex, resolveErr)
		}
		logger.Debug("executing action",
			slog.String("action_type", act.ActionType),
			slog.Int("order_index", act.OrderIndex),
			slog.String("event_id", job.EventID))
		executor, pluginErr := wp.Registry.Get(act.ActionType)
		if pluginErr != nil {
			return pluginErr
		}
		if execErr := executor.Execute(ctx, resolved, job.Payload); execErr != nil {
			return fmt.Errorf("action %s (order %d) failed: %w", act.ActionType, act.OrderIndex, execErr)
		}
	}
	return nil
}

func (wp *WorkerPool) resolveSecrets(ctx context.Context, userID string, config map[string]any) (map[string]any, error) {
	resolved := make(map[string]any, len(config))
	for k, v := range config {
		if strings.HasSuffix(k, "_ref") {
			secretName, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("secret ref %q must be a string", k)
			}
			actualKey := strings.TrimSuffix(k, "_ref")
			value, err := wp.Store.ResolveSecret(ctx, userID, secretName)
			if err != nil {
				return nil, fmt.Errorf("resolve %q: %w", secretName, err)
			}
			resolved[actualKey] = value
		} else {
			resolved[k] = v
		}
	}
	return resolved, nil
}

func (wp *WorkerPool) Shutdown() {
	wp.Logger.Info("Initializing worker pool shutdown")

	if wp.cancel != nil {
		wp.cancel()
	}
	close(wp.JobQueue)
	wp.wg.Wait()
	wp.Logger.Info("Worker pool shutdown complete")
}
