package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/auction-cards/backend/internal/domain"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		log.Printf(
			"%s %s %d %s %d bytes",
			r.Method,
			r.URL.Path,
			wrapped.status,
			duration,
			wrapped.size,
		)
	})
}

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Helper function to send error responses
func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := domain.ErrorResponse(code, message, nil)
	json.NewEncoder(w).Encode(response)
}

// Helper function to send success responses
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := domain.SuccessResponse(data)
	json.NewEncoder(w).Encode(response)
}
