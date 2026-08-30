package http

import (
	"context"
	"kitis-api/internal/database"
	api_errors "kitis-api/internal/errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/jackc/pgx/v5"
)

var valid_source_types = []string{"group", "lecturer", "room"}

func (h *HTTPHandler) parseRequestForm(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		writeError(w, api_errors.ErrHTTPFormInvalid.WithMessage(err))
		return err
	}
	return nil
}

func (h *HTTPHandler) parseFormSourceType(w http.ResponseWriter, r *http.Request) (string, error) {
	source_type := r.PathValue("source_type")
	if !slices.Contains(valid_source_types, source_type) {
		writeError(w, api_errors.ErrParserSourceTypeInvalid)
		return source_type, api_errors.ErrParserSourceTypeInvalid
	}
	return source_type, nil
}
func (h *HTTPHandler) parseFormSource(w http.ResponseWriter, r *http.Request, source_type string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Database.Timeout)
	defer cancel()

	source := r.PathValue("source")
	if db_source, err := database.GetSource(ctx, source_type, source); err != nil {
		var api_err *api_errors.ApiError
		switch err {
		case pgx.ErrNoRows:
			api_err = api_errors.ErrDatabaseSourceNotFound
		case context.DeadlineExceeded:
			api_err = api_errors.ErrDatabaseTimedOut.WithMessageFormat("Database request timed out after %v.", h.config.Database.Timeout)
		default:
			api_errors.ErrDatabaseUnknown.WithMessage(err)
		}
		writeError(w, api_err)
		return source, false, api_err
	} else if db_source.Status == "inactive" {
		return db_source.Name, false, nil
	} else {
		return db_source.Name, true, nil
	}
}
func (h *HTTPHandler) parseFormLesson(w http.ResponseWriter, r *http.Request) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Database.Timeout)
	defer cancel()

	lesson := r.PathValue("lesson")
	if db_lesson, err := database.GetLesson(ctx, lesson); err != nil {
		var api_err *api_errors.ApiError
		switch err {
		case pgx.ErrNoRows:
			api_err = api_errors.ErrDatabaseLessonNotFound
		case context.DeadlineExceeded:
			api_err = api_errors.ErrDatabaseTimedOut.WithMessageFormat("Database request timed out after %v.", h.config.Database.Timeout)
		default:
			api_errors.ErrDatabaseUnknown.WithMessage(err)
		}
		writeError(w, api_err)
		return "", api_err
	} else {
		return db_lesson.Name.String, nil
	}
}

func (h *HTTPHandler) parseFormSubgroup(w http.ResponseWriter, r *http.Request) (int, error) {
	subgroup, err := strconv.Atoi(r.PathValue("subgroup"))
	if err != nil || subgroup < 0 || subgroup > 2 {
		writeError(w, api_errors.ErrParserSubgroupInvalid)
		return 0, api_errors.ErrParserSubgroupInvalid
	}
	return subgroup, nil
}

func (h *HTTPHandler) parseFormTimestamp(w http.ResponseWriter, r *http.Request) (int64, error) {
	timestamp, err := strconv.ParseInt(r.PathValue("timestamp"), 10, 64)
	if err != nil || timestamp < 0 {
		writeError(w, api_errors.ErrParserTimestampInvalid)
		return 0, api_errors.ErrParserTimestampInvalid
	}
	return timestamp, nil
}

func (h *HTTPHandler) parseFormJournalID(w http.ResponseWriter, r *http.Request) (int, error) {
	journal_id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || journal_id < 0 {
		writeError(w, api_errors.ErrParserJournalIDInvalid)
		return 0, api_errors.ErrParserJournalIDInvalid
	}
	return journal_id, nil
}

func (h *HTTPHandler) parseFormSourceID(w http.ResponseWriter, r *http.Request) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Database.Timeout)
	defer cancel()

	source_id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || source_id < 0 {
		writeError(w, api_errors.ErrParserJournalIDInvalid)
		return 0, api_errors.ErrParserJournalIDInvalid
	}

	if _, err := database.GetSourceByID(ctx, source_id); err != nil {
		var api_err *api_errors.ApiError
		switch err {
		case pgx.ErrNoRows:
			api_err = api_errors.ErrDatabaseSourceNotFound
		case context.DeadlineExceeded:
			api_err = api_errors.ErrDatabaseTimedOut.WithMessageFormat("Database request timed out after %v.", h.config.Database.Timeout)
		default:
			api_errors.ErrDatabaseUnknown.WithMessage(err)
		}
		writeError(w, api_err)
		return 0, api_err
	}

	return source_id, nil
}
