package http

import (
	"context"
	"kitis-api/internal/database"
	"net/http"
)

func (h *HTTPHandler) GetSources(w http.ResponseWriter, r *http.Request) {
	if err := h.parseRequestForm(w, r); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.config.Database.Timeout)
	defer cancel()

	// extended result (key: source, value: domain.Source)
	if r.Form.Has("extended") {
		res, err := database.GetSourcesExtended(ctx)
		if err != nil {
			h.handleDatabaseError(w, err)
			return
		}

		writeOK(w, res)
		return
	}

	res, err := database.GetSourcesActive(ctx)
	if err != nil {
		h.handleDatabaseError(w, err)
		return
	}
	writeOK(w, res)
}
func (h *HTTPHandler) GetSourcesByType(w http.ResponseWriter, r *http.Request) {
	if err := h.parseRequestForm(w, r); err != nil {
		return
	}
	source_type, err := h.parseFormSourceType(w, r)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.config.Database.Timeout)
	defer cancel()

	// extended result (key: source, value: domain.Source)
	if r.Form.Has("extended") {
		res, err := database.GetSourcesExtended(ctx)
		if err != nil {
			h.handleDatabaseError(w, err)
			return
		}

		writeOK(w, res[source_type])
		return
	}

	res, err := database.GetSourcesActive(ctx)
	if err != nil {
		h.handleDatabaseError(w, err)
		return
	}
	writeOK(w, res[source_type])
}
