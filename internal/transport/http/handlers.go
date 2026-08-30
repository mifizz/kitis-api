package http

import (
	"context"
	"kitis-api/internal/config"
	"kitis-api/internal/errors"
	api_errors "kitis-api/internal/errors"
	"kitis-api/internal/parser"
	"kitis-api/internal/watcher"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type HTTPHandler struct {
	config  *config.Config
	parser  *parser.Parser
	watcher *watcher.Watcher
}

func NewHTTPHandler(config *config.Config, parser *parser.Parser, watcher *watcher.Watcher) *HTTPHandler {
	return &HTTPHandler{config: config, parser: parser, watcher: watcher}
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	writeError(w, errors.ErrHTTPMethodNotFound)
}

func (h *HTTPHandler) handleHTTPError(w http.ResponseWriter, err error) {
	switch err {
	case context.DeadlineExceeded:
		writeError(w, api_errors.ErrHTTPTimedOut.WithMessageFormat("HTTP request timed out %v seconds.", h.config.Parser.Timeout))
	default:
		writeError(w, api_errors.ErrHTTPUnknown.WithMessage(err))
	}
}

func (h *HTTPHandler) handleDatabaseError(w http.ResponseWriter, err error) {
	switch err {
	case pgx.ErrNoRows:
		writeError(w, api_errors.ErrDatabaseSourcesNotFound)
	case context.DeadlineExceeded:
		writeError(w, api_errors.ErrDatabaseTimedOut.WithMessageFormat("Database timed out after %v.", h.config.Database.Timeout))
	default:
		writeError(w, api_errors.ErrDatabaseUnknown.WithMessage(err))
	}
}
