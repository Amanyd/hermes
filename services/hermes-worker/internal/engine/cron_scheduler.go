package engine

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/cronutil"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/store"
	"github.com/google/uuid"
)

type CronStore interface {
	GetCronRelaysDue(ctx context.Context) ([]store.CronRelay, error)
	UpdateRelayNextRun(ctx context.Context, relayID string, nextRunAt *time.Time) error
}

type CronScheduler struct {
	store    CronStore
	jobQueue chan Job
	logger   *slog.Logger
	ticker   *time.Ticker
	done     chan struct{}
}

func NewCronScheduler(store CronStore, jobQueue chan Job, logger *slog.Logger) *CronScheduler {
	return &CronScheduler{
		store:    store,
		jobQueue: jobQueue,
		logger:   logger,
		done:     make(chan struct{}),
	}
}

func (cs *CronScheduler) Start(ctx context.Context) {
	cs.ticker = time.NewTicker(30 * time.Second)
	cs.logger.Info("cron scheduler started")

	cs.tick(ctx)

	go func() {
		for {
			select {
			case <-cs.done:
				cs.ticker.Stop()
				cs.logger.Info("cron scheduler stopped")
				return
			case <-ctx.Done():
				cs.ticker.Stop()
				cs.logger.Info("cron scheduler context cancelled")
				return
			case <-cs.ticker.C:
				cs.tick(ctx)
			}
		}
	}()
}

func (cs *CronScheduler) Stop() {
	close(cs.done)
}

func (cs *CronScheduler) tick(ctx context.Context) {
	due, err := cs.store.GetCronRelaysDue(ctx)
	if err != nil {
		cs.logger.Error("failed to fetch due cron relays", slog.String("error", err.Error()))
		return
	}

	for _, relay := range due {
		cs.logger.Info("firing cron relay", slog.String("relay_id", relay.ID))

		payload, _ := json.Marshal(map[string]any{
			"trigger":  "cron",
			"fired_at": time.Now().UTC().Format(time.RFC3339),
			"relay_id": relay.ID,
		})

		job := Job{
			RelayID: relay.ID,
			EventID: uuid.New().String(),
			Payload: payload,
			MsgAck:  func(bool) {},
		}

		select {
		case cs.jobQueue <- job:
		default:
			cs.logger.Warn("job queue full, cron relay skipped", slog.String("relay_id", relay.ID))
			continue
		}

		schedule, _ := relay.TriggerConfig["schedule"].(string)
		nextRun, err := cronutil.ComputeNextRun(schedule, time.Now())
		if err != nil {
			cs.logger.Error("failed to compute next run",
				slog.String("relay_id", relay.ID),
				slog.String("error", err.Error()))
			continue
		}
		if err := cs.store.UpdateRelayNextRun(ctx, relay.ID, &nextRun); err != nil {
			cs.logger.Error("failed to update next_run_at",
				slog.String("relay_id", relay.ID),
				slog.String("error", err.Error()))
		}
	}
}
