package middleware

import (
	"net/http"
	"strings"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) for the API
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if the origin is in the allowed list
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin || allowedOrigin == "*" {
					allowed = true
					break
				}
				// Support wildcard subdomains like *.trycloudflare.com
				if strings.HasPrefix(allowedOrigin, "*.") {
					domain := strings.TrimPrefix(allowedOrigin, "*.")
					if strings.HasSuffix(origin, domain) {
						allowed = true
						break
					}
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetAllowedOrigins returns the list of allowed origins from environment or defaults
func GetAllowedOrigins() []string {
	// You can extend this to read from environment variables
	return []string{
		"https://solid-guide-974w6599qrjgfpr9j-5174.app.github.dev",
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:8080",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:3001",
		"*.trycloudflare.com", // Support any Cloudflare tunnel
		"*.ngrok.io",          // Support ngrok tunnels
		"*.localhost.run",     // Support localhost.run tunnels
	}
}
