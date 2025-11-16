package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ImamSR/go-books-api/internal/books"
	"github.com/ImamSR/go-books-api/internal/db"
    "github.com/ImamSR/go-books-api/internal/httpx"
	"github.com/ImamSR/go-books-api/internal/auth"
	"github.com/ImamSR/go-books-api/internal/users"
)

func main() {
	ctx := context.Background()
	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatal(err) // untuk auth & persistence kita butuh DB
	}

	// books
	bookStore := books.NewPGStore(pool)
	bh := books.NewHandler(bookStore)

	// users
	userRepo := users.NewPGRepo(pool)
	tokenGen := auth.NewTokenGenerator(auth.MustJWTSecret())
	uh := users.NewHandler(userRepo, tokenGen)

	router := httpx.NewRouter(bh, uh)

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" { addr = ":" + v }

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("listening on", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}