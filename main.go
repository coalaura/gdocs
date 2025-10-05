package main

import (
	"net/http"

	"github.com/coalaura/plain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var log = plain.New(plain.WithDate(plain.RFC3339Local))

func main() {
	go cleanupLoop()

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(log.Middleware())

	r.Get("/{doc}/{page}", handleDoc)

	log.Println("Listening at http://localhost:4176/")
	http.ListenAndServe(":4176", r)
}
