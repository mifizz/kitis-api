package watcher

import (
	"context"
	"kitis-api/internal/database"
	api_errors "kitis-api/internal/errors"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

func (w *Watcher) saveSchedule(ctx context.Context, source_type, source string) error {
	res, err := w.parser.GetSchedule(ctx, source_type, source)
	if err != nil {
		return err
	}
	if equal, err := compareSchedule(ctx, res); equal {
		return api_errors.ErrJSONNotModified
	} else if err != nil && err != pgx.ErrNoRows {
		slog.Error("failed to compare schedules", "source", source, "error", err)
		return err
	}
	return database.AddSchedule(ctx, res)
}

func (w *Watcher) saveRecords(ctx context.Context, source_type, source string) error {
	res, err := w.parser.GetRecords(ctx, source_type, source)
	if err != nil {
		return err
	}
	return database.AddRecords(ctx, res)
}

func (w *Watcher) SaveRegular(ctx context.Context) {
	slog.Info("saving regular...")
	time_start := time.Now()
	var mutex sync.Mutex
	var updated_sources = map[string][]string{
		"group":    {},
		"lecturer": {},
		"room":     {},
	}
	var g errgroup.Group
	g.SetLimit(10)

	for source_type, sources := range w.parser.GetSources() {
		for source := range sources {
			g.Go(func() error {
				if _, err := database.GetSource(ctx, source_type, source); err != nil {
					database.AddSource(ctx, source_type, source)
					slog.Debug("added new source", "source", source)
				}

				slog.Debug("saving regular", "source", source)
				// schedule
				err := w.saveSchedule(ctx, source_type, source)
				if err != nil {
					slog.Warn("[schedule] failed to save", "source", source, "error", err)
				} else {
					mutex.Lock()
					updated_sources[source_type] = append(updated_sources[source_type], source)
					mutex.Unlock()
				}

				// records
				err = w.saveRecords(ctx, source_type, source)
				if err != nil {
					slog.Warn("[records] failed to save", "source", source, "error", err)
				}

				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		slog.Error("watcher err", "error", err)
	}

	w.PushMessage(map[string]any{
		"schedule": updated_sources,
	})

	slog.Info("done saving regular", "elapsed", time.Since(time_start))
}
