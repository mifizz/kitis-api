package watcher

import (
	"context"
	"io"
	"kitis-api/internal/database"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

func (w *Watcher) saveJournal(ctx context.Context, journal_id int) error {
	res, err := w.parser.GetJournal(ctx, journal_id)
	if err != nil {
		return err
	}
	return database.AddJournal(ctx, res)
}

func (w *Watcher) SaveAllJournals(ctx context.Context) {
	slog.Info("saving journals...")
	time_start := time.Now()
	var g errgroup.Group
	g.SetLimit(10)

	for group := range w.parser.GetSources()["group"] {
		g.Go(func() error {
			recs, err := database.GetRecords(ctx, "group", group)
			if err == pgx.ErrNoRows {
				slog.Debug("no records", "group", group)
				return nil
			} else if err != nil {
				slog.Error("database error", "group", group, "error", err)
				return nil
			}
			for _, record := range recs.Rows {
				slog.Debug("saving journal", "group", group, "id", record.JournalId)
				err := w.saveJournal(ctx, record.JournalId)
				if err == io.EOF {
					slog.Warn("[journal] EOF - skipping")
				} else if err != nil {
					slog.Warn("[journal] failed to save", "error", err)
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		slog.Error("watcher err", "error", err)
	}

	slog.Info("done saving journals", "elapsed", time.Since(time_start))
}
