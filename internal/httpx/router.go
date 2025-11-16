package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ImamSR/go-books-api/internal/books"
	"github.com/ImamSR/go-books-api/internal/auth"
	"github.com/ImamSR/go-books-api/internal/users"
)

func NewRouter(bh *books.Handler, uh *users.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return CommonMiddlewares(next) })

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// auth
	r.Route("/auth", func(ar chi.Router) {
		ar.Post("/register", uh.Register)
		ar.Post("/login", uh.Login)
	})

	// books (GET publik)
	r.Get("/books", bh.List)
	r.Get("/books/{id}", bh.Detail)

	// books (write: require JWT + roles)
	sec := auth.MustJWTSecret()
	protected := chi.NewRouter()
	protected.Use(auth.AuthJWT(sec))
	protected.With(auth.RequireRoles("editor", "admin")).Post("/books", bh.Create)
	protected.With(auth.RequireRoles("editor", "admin")).Put("/books/{id}", bh.Update)
	protected.With(auth.RequireRoles("admin")).Delete("/books/{id}", bh.Delete)
	r.Mount("/", protected)

	return r
}
