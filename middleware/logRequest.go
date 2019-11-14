package middleware

import (
	"net/http"
	"log"
)

// LogRequest
// Automatically log all relevant information about a request.
func LogRequest(next http.Handler) (http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.String(), r.Host, r.RemoteAddr, r.UserAgent())
		next.ServeHTTP(w, r)
	})
}