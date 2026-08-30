package http

import (
	"context"
	"kitis-api/internal/database"
	api_errors "kitis-api/internal/errors"
	"net/http"
)

func (h *HTTPHandler) GetJournalById(w http.ResponseWriter, r *http.Request) {
	journal_id, err := h.parseFormJournalID(w, r)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.config.Parser.Timeout)
	defer cancel()

	res, err := h.parser.GetJournal(ctx, journal_id)
	if err != nil {
		switch err {
		case context.DeadlineExceeded:
			writeError(w, api_errors.ErrHTTPTimedOut.WithMessageFormat("HTTP request timed out %v seconds.", h.config.Parser.Timeout))
		case api_errors.ErrParserJournalNotFound:
			writeError(w, api_errors.ErrParserJournalNotFound)
		case api_errors.ErrParserPageNotFound:
			writeError(w, api_errors.ErrParserPageNotFound)
		default:
			writeError(w, api_errors.ErrHTTPUnknown.WithMessage(err))
		}
		return
	}

	writeOK(w, res)
}

func (h *HTTPHandler) GetJournal(w http.ResponseWriter, r *http.Request) {
	source, _, err := h.parseFormSource(w, r, "group")
	if err != nil {
		return
	}
	subgroup, err := h.parseFormSubgroup(w, r)
	if err != nil {
		return
	}
	lesson, err := h.parseFormLesson(w, r)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.config.Database.Timeout)
	defer cancel()

	res, err := database.GetJournal(ctx, source, subgroup, lesson)
	if err != nil {
		h.handleDatabaseError(w, err)
		return
	}

	writeOK(w, res)
}
