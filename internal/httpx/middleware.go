package httpx

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func CommonMiddlewares(next http.Handler) http.Handler {
	chain := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
	// You can add chi middlewares if you want:
	_ = middleware.RequestID
	return chain
}
