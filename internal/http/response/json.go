package response

import (
	"encoding/json"
	"net/http"
)

// Success wraps payloads in the common envelope expected by clients.
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	writeJSON(w, statusCode, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

// Error wraps error responses consistently.
func Error(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]interface{}{
		"status":  "error",
		"message": message,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
