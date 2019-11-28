package middleware

import(
	"net/http"
	"strings"
)

// CORSConfig is a struct holding all relevant
// information for a proper CORS configuration.
// Fields correspond to CORS headers.
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders string
	MaxAge 		   string
}

// GetOriginsStr returns the comma delimited
// string array of a CORSConfig struct's
// AllowedOrigins field.
func (c *CORSConfig) GetOriginsStr() string {
	return strings.Join(c.AllowedOrigins[:], ",")
}

// GetMethodsStr returns the comma delimited string
// array of a CORSConfig struct's AllowedMethods field.
func (c *CORSConfig) GetMethodsStr() string {
	return strings.Join(c.AllowedMethods[:], ",")
}

// WithCORSConfig is middleware that handles your CORS preflight requests quickly
// and effectively with the supplied configuration. It is not verbose. To enable
// verbosity, please wrap it with some sort of request logging middleware.
func WithCORSConfig(next http.Handler, c CORSConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// handle preflight request.
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", c.GetOriginsStr())
			w.Header().Set("Access-Control-Allow-Methods", c.GetMethodsStr())
			w.Header().Set("Access-Control-Allow-Headers", c.AllowedHeaders)
			w.Header().Set("Access-Control-Max-Age", c.MaxAge)
			w.WriteHeader(http.StatusOK)
			return
		}

		// serve actual request.
		next.ServeHTTP(w, r)
	})
}

// WithCORS is middleware that handles your CORS preflight requests quickly
// and effectively using default settings. It is not verbose. To enable
// verbosity, please wrap it with some sort of request logging middleware.
func WithCORS(next http.Handler) http.Handler {
	// by default we allow all methods, all origins,
	// and Content-Type headers.
	defaultCfg := CORSConfig{
		AllowedHeaders: "Content-Type",
		AllowedMethods: []string{
						http.MethodConnect,
						http.MethodDelete,
						http.MethodGet,
						http.MethodHead,
						http.MethodOptions,
						http.MethodPatch,
						http.MethodPost,
						http.MethodPut,
						http.MethodTrace,
						},
		AllowedOrigins: []string{"*"},
		MaxAge: "72800",
		}
	
	return WithCORSConfig(next, defaultCfg)
}