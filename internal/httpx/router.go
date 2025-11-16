package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/your-username/go-books-api/internal/books"
)

func NewRouter(bh *books.Handler) http.Handler {
	r := chi.NewRouter()
	// attach logging/other middlewares
	r.Use(func(next http.Handler) http.Handler { return CommonMiddlewares(next) })

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/books", func(r chi.Router) {
		r.Get("/", bh.List)
		r.Post("/", bh.Create)
		r.Get("/{id}", bh.Detail)
		r.Put("/{id}", bh.Update)
		r.Delete("/{id}", bh.Delete)
	})

	return r
}
