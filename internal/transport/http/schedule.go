package http

import (
	"context"
	"kitis-api/internal/database"
	"kitis-api/internal/domain"
	api_errors "kitis-api/internal/errors"

	"net/http"
)

func (h *HTTPHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	// Check if form parses correctly (idk how it possibly would not)
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

		var res domain.Schedule
		res, err := h.parser.GetSchedule(ctx, source_type, source)
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

	res, err := database.GetScheduleNewest(ctx, source_type, source)
	if err != nil {
		h.handleDatabaseError(w, err)
		return
	}

	writeOK(w, res)
}
func (h *HTTPHandler) GetScheduleByTimestamp(w http.ResponseWriter, r *http.Request) {
	source_type, err := h.parseFormSourceType(w, r)
	if err != nil {
		return
	}
	source, _, err := h.parseFormSource(w, r, source_type)
	if err != nil {
		return
	}
	timestamp, err := h.parseFormTimestamp(w, r)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.config.Database.Timeout)
	defer cancel()

	res, err := database.GetScheduleByTimestamp(ctx, source_type, source, timestamp)
	if err != nil {
		h.handleDatabaseError(w, err)
		return
	}

	writeOK(w, res)
}

func (h *HTTPHandler) GetScheduleVersions(w http.ResponseWriter, r *http.Request) {
	source_type, err := h.parseFormSourceType(w, r)
	if err != nil {
		return
	}
	source, _, err := h.parseFormSource(w, r, source_type)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.config.Database.Timeout)
	defer cancel()

	res, err := database.GetScheduleVersions(ctx, source_type, source)
	if len(res) < 1 {
		writeError(w, api_errors.ErrDatabaseScheduleVersionsNotFound)
		return
	}
	if err != nil {
		h.handleDatabaseError(w, err)
		return
	}

	writeOK(w, res)
}

func (h *HTTPHandler) GetScheduleByID(w http.ResponseWriter, r *http.Request) {
	source_id, err := h.parseFormSourceID(w, r)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.config.Database.Timeout)
	defer cancel()

	res, err := database.GetScheduleByIDNewest(ctx, source_id)
	if err != nil {
		h.handleDatabaseError(w, err)
		return
	}

	writeOK(w, res)
}
