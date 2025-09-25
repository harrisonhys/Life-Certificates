package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// BasicAuth protects endpoints using HTTP Basic authentication.
func BasicAuth(username, password string) func(http.Handler) http.Handler {
	realm := "Restricted"
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !validateBasicAuth(auth, username, password) {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", realm))
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func validateBasicAuth(header, username, password string) bool {
	if !strings.HasPrefix(header, "Basic ") {
		return false
	}

	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(header, "Basic "))
	if err != nil {
		return false
	}

	parts := strings.SplitN(string(payload), ":", 2)
	if len(parts) != 2 {
		return false
	}

	return parts[0] == username && parts[1] == password
}
