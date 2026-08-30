package watcher

import (
	"context"
	"errors"
	"kitis-api/internal/database"
	"log/slog"
	"os"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
)

func (w *Watcher) checkRegularUpdates(ctx context.Context) (bool, error) {
	mt_present, err := w.parser.GetModifiedTime(ctx)
	if err != nil {
		slog.Error("failed to parse modified time", "error", err)
		if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, context.DeadlineExceeded) {
			// maybe add something here later
			return false, err
		}
		return false, err
	}

	if _, err = database.CheckModifiedTime(ctx, mt_present); err != nil {
		return true, err
	}

	return false, nil
}
func (w *Watcher) checkJournalUpdates(ctx context.Context) (bool, error) {
	groups := w.parser.GetSources()["group"]
	var group string
	for source := range groups {
		group = source

		recs, err := w.parser.GetRecords(ctx, "group", group)
		if err != nil {
			slog.Error("failed to fetch records", "error", err)
			return false, err
		} else if len(recs.Rows) == 0 {
			// skip to next group
			continue
		}

		jrn, err := w.parser.GetJournal(ctx, recs.Rows[0].JournalId)
		if err != nil {
			slog.Error("failed to fetch journal", "error", err)
			return false, err
		} else if len(jrn.Rows) == 0 {
			// skip to next group
			continue
		}

		jrn_db, err := database.GetJournal(ctx, jrn.Group, jrn.Subgroup, jrn.Lesson)
		if err != nil {
			slog.Error("failed to load database journal", "error", err)
			return false, err
		}

		md_present, err := time.Parse("02.01.2006", jrn.ModifiedDate)
		if err != nil {
			slog.Error("can't parse modified date", "error", err, "src", jrn.ModifiedDate)
			return false, err
		}
		md_db, err := time.Parse("02.01.2006", jrn_db.ModifiedDate)
		if err != nil {
			slog.Error("can't parse modified date", "error", err, "src", jrn_db.ModifiedDate)
			return false, err
		}

		if md_present.Unix() > md_db.Unix() {
			return true, nil
		}
		break
	}
	return false, nil
}

func (w *Watcher) checkModifiedTime(ctx context.Context) (bool, error) {
	mt_actual, err := w.parser.GetModifiedTime(ctx)
	if os.IsTimeout(err) || errors.Is(err, syscall.ECONNRESET) {
		return false, err
	} else if err != nil {
		return true, err
	}

	if _, err = database.CheckModifiedTime(ctx, mt_actual); err != nil {
		return true, err
	}
	return false, nil
}

func (w *Watcher) checkJournalModifiedDate(ctx context.Context) (bool, error) {
	sources := w.parser.GetSources()["group"]
	// TODO: add source caching here if updated
	var group string
	for source := range sources {
		group = source
		recs, err := w.parser.GetRecords(ctx, "group", group)
		if os.IsTimeout(err) {
			return false, err
		} else if err != nil || len(recs.Rows) < 1 {
			continue
		}

		jrn, err := w.parser.GetJournal(ctx, recs.Rows[0].JournalId)
		if os.IsTimeout(err) {
			return false, err
		} else if err != nil || len(jrn.Rows) < 1 {
			continue
		}
		md_actual, err := time.Parse("02.01.2006", jrn.ModifiedDate)
		if err != nil {
			slog.Error("failed to parse modified date", "error", err, "group", group)
			continue
		}

		jrn_db, err := database.GetJournal(ctx, jrn.Group, jrn.Subgroup, jrn.Lesson)
		if err == pgx.ErrNoRows {
			return true, err
		} else if err != nil {
			slog.Error("database error", "error", err, "group", group)
			continue
		}
		md_db, err := time.Parse("02.01.2006", jrn_db.ModifiedDate)
		if err != nil {
			slog.Error("failed to parse modified date", "error", err, "group", group)
			continue
		}

		if md_actual.Unix() > md_db.Unix() {
			return true, nil
		}
		break
	}
	return false, nil
}

func (w *Watcher) Check(ctx context.Context) {
	changed, err := w.checkModifiedTime(ctx)
	if os.IsTimeout(err) {
		slog.Debug("host timed out, skipping update")
		return
	} else if errors.Is(err, syscall.ECONNRESET) {
		slog.Error("skipping watcher check, site is down", "error", err)
	} else if err != nil && err != pgx.ErrNoRows {
		slog.Error("watcher check error", "error", err)
	}
	if changed {
		w.SaveRegular(ctx)
		return
	}

	slog.Debug("no schedule/records updates")
}

func (w *Watcher) CheckJournals(ctx context.Context) {
	changed, err := w.checkJournalModifiedDate(ctx)
	if os.IsTimeout(err) {
		slog.Debug("host timed out, skipping journals update")
		return
	} else if err != nil && err != pgx.ErrNoRows {
		slog.Error("watcher check error", "error", err)
	}
	if changed {
		w.SaveAllJournals(ctx)
		return
	}

	slog.Debug("no journal updates")
}
