package http

import (
	"context"
	"kitis-api/internal/database"
	"kitis-api/internal/domain"
	api_errors "kitis-api/internal/errors"
	"net/http"
)

func (h *HTTPHandler) GetRecords(w http.ResponseWriter, r *http.Request) {
	if err := h.parseRequestForm(w, r); err != nil {
		return
	}
	source_type, err := h.parseFormSourceType(w, r)
	if err != nil {
		return
	}
	source, source_active, err := h.parseFormSource(w, r, source_type)
	if err != nil {
		return
	}
	force := r.Form.Has("force")

	// fetch from server
	if force {
		if !source_active {
			writeError(w, api_errors.ErrParserSourceOutdated)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), h.config.Parser.Timeout)
		defer cancel()

		var res domain.Records
		res, err := h.parser.GetRecords(ctx, source_type, source)
		if err != nil {
			h.handleHTTPError(w, err)
			return
		}

		writeOK(w, res)
		return
	}

	// use database
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Database.Timeout)
	defer cancel()

	res, err := database.GetRecords(ctx, source_type, source)
	if err != nil {
		h.handleDatabaseError(w, err)
		return
	}

	writeOK(w, res)
}
