package httputils

import (
	"duck/common"
	"encoding/json"
	"net/http"
)

// ResponseWriter wraps http.ResponseWriter with logging
type ResponseWriter struct {
	logger *common.Logger
}

// NewResponseWriter creates a new response writer
func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		logger: common.NewLogger("response-writer"),
	}
}

// WriteJSON writes a JSON response
func (rw *ResponseWriter) WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	rw.logger.Info("Writing JSON response")
	return json.NewEncoder(w).Encode(data)
}
