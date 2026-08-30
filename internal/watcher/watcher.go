package watcher

import (
	"context"
	"kitis-api/internal/config"
	"kitis-api/internal/parser"
	"log/slog"
	"sync"
	"time"
)

type Watcher struct {
	config    *config.WatcherConfig
	parser    *parser.Parser
	mutex     sync.Mutex
	Broadcast chan []byte
}

func NewWatcher(config *config.WatcherConfig, parser *parser.Parser) *Watcher {
	return &Watcher{config: config, parser: parser, Broadcast: make(chan []byte, 256)}
}

func (w *Watcher) UpdateRegular(ctx context.Context) {
	slog.Debug("checking regular updates...")
	if changed, err := w.checkRegularUpdates(ctx); changed {
		if err := w.parser.UpdateSources(ctx); err != nil {
			slog.Warn("skipping regular update")
			w.mutex.Unlock()
			return
		}
		w.SaveRegular(ctx)
	} else if err != nil {
		slog.Warn("skipping regular update")
	} else {
		slog.Debug("no regular updates")
	}
}
func (w *Watcher) UpdateJournals(ctx context.Context) {
	if err := w.parser.UpdateSources(ctx); err != nil {
		slog.Warn("skipping journals update")
		w.mutex.Unlock()
		return
	}
	slog.Debug("checking journal updates...")
	if changed, err := w.checkJournalUpdates(ctx); changed {
		w.SaveAllJournals(ctx)
	} else if err != nil {
		slog.Warn("skipping journals update")
	} else {
		slog.Debug("no journal updates")
	}
}

func (w *Watcher) Start(ctx context.Context) {
	ticker := time.NewTicker(w.config.Wait)
	defer ticker.Stop()
	// source_ticker := time.NewTicker(w.config.SourceWait)
	// defer source_ticker.Stop()
	journal_ticker := time.NewTicker(w.config.JournalWait)
	defer journal_ticker.Stop()

	slog.Info("watcher started")

	w.UpdateRegular(ctx)
	w.UpdateJournals(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("watcher stopped")
			return
		case <-ticker.C:
			w.mutex.Lock()
			w.UpdateRegular(ctx)
			w.mutex.Unlock()
		case <-journal_ticker.C:
			w.mutex.Lock()
			w.UpdateJournals(ctx)
			w.mutex.Unlock()
		}
	}
}
