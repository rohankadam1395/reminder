package worker

import (
	"context"
	"reminder/service"
	"time"

	kitlog "github.com/go-kit/log"
)

type Config struct {
	PollInterval time.Duration
	BatchSize    int
}

type Worker struct {
	svc    service.ReminderService
	logger kitlog.Logger
	cfg    Config
}

func New(svc service.ReminderService, logger kitlog.Logger, cfg Config) *Worker {
	if cfg.BatchSize < 1 {
		cfg.BatchSize = 50
	}
	return &Worker{svc: svc, logger: logger, cfg: cfg}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()

	w.logger.Log("msg", "worker started", "interval", w.cfg.PollInterval)

	for {
		select {
		case <-ctx.Done():
			w.logger.Log("msg", "worker stopped")
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	reminders, err := w.svc.ListDueReminders(ctx, time.Now().UTC(), w.cfg.BatchSize)
	if err != nil {
		w.logger.Log("msg", "failed to list due reminders", "error", err)
		return
	}

	for _, r := range reminders {
		w.logger.Log("msg", "sending reminder", "id", r.ID, "message", r.Message)
		if err := w.svc.MarkReminderNotified(ctx, r.ID); err != nil {
			w.logger.Log("msg", "failed to mark reminder notified", "id", r.ID, "error", err)
		}
	}
}
