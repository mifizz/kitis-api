package errors

import (
	"fmt"
	"net/http"
)

type ApiError struct {
	HTTPStatus int
	Code       string
	Message    string
}

func (e *ApiError) Error() string {
	return e.Message
}
func (e *ApiError) WithMessageFormat(format string, args ...any) *ApiError {
	// copy error
	new_err := *e
	new_err.Message = fmt.Sprintf(format, args...)
	return &new_err
}
func (e *ApiError) WithMessage(args ...any) *ApiError {
	// copy error
	new_err := *e
	new_err.Message = fmt.Sprint(args...)
	return &new_err
}

var (
	// --- Database ---

	ErrDatabaseSourceNotFound = &ApiError{
		HTTPStatus: http.StatusNotFound,
		Code:       "database.source.not_found",
		Message:    "Source not found in database.",
	}
	ErrDatabaseSourcesNotFound = &ApiError{
		HTTPStatus: http.StatusNotFound,
		Code:       "database.sources.not_found",
		Message:    "Sources not found in database.",
	}
	ErrDatabaseLessonNotFound = &ApiError{
		HTTPStatus: http.StatusNotFound,
		Code:       "database.lesson.not_found",
		Message:    "Lesson not found in database.",
	}
	ErrDatabaseScheduleNotFound = &ApiError{
		HTTPStatus: http.StatusNotFound,
		Code:       "database.schedule.not_found",
		Message:    "Schedule not found in database.",
	}
	ErrDatabaseScheduleVersionsNotFound = &ApiError{
		HTTPStatus: http.StatusNotFound,
		Code:       "database.schedule.no_versions",
		Message:    "Schedule versions not found in database.",
	}
	ErrDatabaseRecordsNotFound = &ApiError{
		HTTPStatus: http.StatusNotFound,
		Code:       "database.records.not_found",
		Message:    "Records not found in database.",
	}
	ErrDatabaseJournalNotFound = &ApiError{
		HTTPStatus: http.StatusNotFound,
		Code:       "database.journal.not_found",
		Message:    "Journal not found in database.",
	}

	ErrDatabaseTimedOut = &ApiError{
		HTTPStatus: http.StatusGatewayTimeout,
		Code:       "database.timed_out",
		Message:    "Database timed out.",
	}
	ErrDatabaseUnknown = &ApiError{
		HTTPStatus: http.StatusInternalServerError,
		Code:       "database.unknown_error",
		Message:    "Unknown database error.",
	}

	// --- Parser ---

	ErrParserSourceTypeInvalid = &ApiError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "parser.source_type.invalid",
		Message:    "Invalid source type. Must be either 'group', 'lecturer' or 'room'.",
	}
	ErrParserSourceOutdated = &ApiError{
		HTTPStatus: http.StatusGone,
		Code:       "parser.source.outdated",
		Message:    "Source is outdated.",
	}
	ErrParserSourceNotFound = &ApiError{
		HTTPStatus: http.StatusNotFound,
		Code:       "parser.source.not_found",
		Message:    "Source not found.",
	}
	ErrParserSourceIDInvalid = &ApiError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "parser.source.invalid_id",
		Message:    "Invalid source ID. Must be unsigned integer.",
	}

	ErrParserTimestampInvalid = &ApiError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "parser.timestamp.invalid",
		Message:    "Invalid timestamp. Must be unsigned 64-bit integer.",
	}
	ErrParserJournalIDInvalid = &ApiError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "parser.journal.invalid_id",
		Message:    "Invalid journal ID. Must be unsigned integer.",
	}
	ErrParserSubgroupInvalid = &ApiError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "parser.subgroup.invalid",
		Message:    "Invalid subgroup. Must be either 0, 1 or 2, where 0 is full group.",
	}

	ErrParserNotImplemented = &ApiError{
		HTTPStatus: http.StatusNotImplemented,
		Code:       "parser.not_implemented",
		Message:    "Could not parse data.",
	}

	ErrParserJournalNotFound = &ApiError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "parser.journal.not_found",
		Message:    "Journal not found.",
	}

	ErrParserPageNotFound = &ApiError{
		HTTPStatus: http.StatusNotFound,
		Code:       "parser.page.not_found",
		Message:    "Page not found on server.",
	}
	ErrParserUnknown = &ApiError{
		HTTPStatus: http.StatusInternalServerError,
		Code:       "parser.unknown_error",
		Message:    "Unknown parser error.",
	}

	// --- HTTP ---

	ErrHTTPFormInvalid = &ApiError{
		HTTPStatus: http.StatusNotAcceptable,
		Code:       "http.form.invalid",
		Message:    "Invalid HTTP form.",
	}
	ErrHTTPFormInvalidParameter = &ApiError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "http.form.invalid_parameter",
		Message:    "Invalid HTTP form parameter.",
	}
	ErrHTTPFormWrongParameterType = &ApiError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "http.form.wrong_parameter_type",
		Message:    "Wrong HTTP form parameter type.",
	}
	ErrHTTPTimedOut = &ApiError{
		HTTPStatus: http.StatusGatewayTimeout,
		Code:       "http.timed_out",
		Message:    "Timed out.",
	}
	ErrHTTPMethodNotFound = &ApiError{
		HTTPStatus: http.StatusNotFound,
		Code:       "http.method.not_found",
		Message:    "Method not found.",
	}
	ErrHTTPUnknown = &ApiError{
		HTTPStatus: http.StatusInternalServerError,
		Code:       "http.unknown_error",
		Message:    "Unknown HTTP error.",
	}

	// --- Other ---

	ErrJSONNotModified = &ApiError{
		HTTPStatus: http.StatusNotModified,
		Code:       "json.not_modified",
		Message:    "Not modified.",
	}
	ErrJSONMarshalFailed = &ApiError{
		HTTPStatus: http.StatusInternalServerError,
		Code:       "json.marshal.failed",
		Message:    "Failed to marshal response JSON.",
	}
)
