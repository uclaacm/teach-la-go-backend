package middleware

import (
	"log"
	"net/http"
)

// LogRequest automatically logs all relevant information about a request object
// before forwarding it to the next Handler provided as its argument.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.String(), r.Host, r.RemoteAddr, r.UserAgent())
		next.ServeHTTP(w, r)
	})
}
