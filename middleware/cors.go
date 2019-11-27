package middleware

import "net/http"

// WithCORS is middleware that handles your CORS preflight requests quickly
// and effectively. It is not verbose. To enable verbosity, please wrap it
// with some sort of request logging middleware.
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// handle preflight request.
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "72800")
			w.WriteHeader(http.StatusNotImplemented)
			return
		}

		// serve actual request.
		next.ServeHTTP(w, r)
	})
}