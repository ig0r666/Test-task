package middleware

import (
	"net/http"
	"strings"
	"testtask/limiter/core"
)

func Rate(next http.HandlerFunc, rate core.RateLimiter, db core.RateLimiterDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := strings.Split(r.RemoteAddr, ":")

		// Узнаем, есть ли у пользователя токены
		allow, err := rate.AllowClientRequest(r.Context(), clientID[0], db)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !allow {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}
