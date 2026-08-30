package http

import (
	api_errors "kitis-api/internal/errors"

	"encoding/json"
	"fmt"
	"net/http"
)

type ApiErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func writeJSON(w http.ResponseWriter, status int, obj any) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(obj)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"message":"Failed to marshal response JSON: %s","code":"%s"}`, err.Error(), api_errors.ErrJSONMarshalFailed.Code)
		return
	}
	w.WriteHeader(status)
	w.Write(data)
}

func writeError(w http.ResponseWriter, err *api_errors.ApiError) {
	writeJSON(w, err.HTTPStatus, ApiErrorResponse{
		Message: err.Message,
		Code:    err.Code,
	})
}

func writeOK(w http.ResponseWriter, obj any) {
	writeJSON(w, http.StatusOK, obj)
}
