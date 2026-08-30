package http

import (
	"kitis-api/internal/domain"
	"net/http"
	"strconv"
)

func (h *HTTPHandler) GetBells(w http.ResponseWriter, r *http.Request) {
	var bells domain.Bells
	bells.General = map[string]string{}
	bells.Monday = map[string]string{}
	for i, b := range h.config.Parser.Bells.General {
		if i < 1 || i > 6 {
			continue
		}
		bells.General[strconv.Itoa(i)] = b
	}
	for i, b := range h.config.Parser.Bells.Monday {
		if (i < 1 && i != -1) || i > 7 {
			continue
		}
		bells.Monday[strconv.Itoa(i)] = b
	}

	writeOK(w, bells)
}
