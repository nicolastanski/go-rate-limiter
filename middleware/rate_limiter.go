package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/nicolastanski/go-rate-limiter/internal"
)

func RateLimitMiddleware(client *redis.Client) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()
			ip := strings.Split(r.RemoteAddr, ":")[0]
			token := r.Header.Get("API_KEY")

			key := "ip"
			requestValue := ip

			if token != "" {
				key = "token"
				requestValue = "token:" + token
			}

			limiter := internal.NewRateLimiter(client, key)

			allowed, err := limiter.Allow(ctx, key, requestValue)

			if err != nil {
				fmt.Println(err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
