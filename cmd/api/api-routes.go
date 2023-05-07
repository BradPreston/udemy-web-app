package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	// register middleware
	mux.Use(middleware.Recoverer)
	mux.Use(app.enableCORS)

	// authentication routes
	mux.Post("/auth", app.authenticate)
	mux.Post("/refresh-token", app.refresh)

	// protected routes
	mux.Route("/users", func(mux chi.Router) {
		// use auth middleware
		mux.Use(app.authRequired)
		mux.Get("/", app.allUsers)
		mux.Post("/", app.insertUser)
		mux.Get("/{userID}", app.getUser)
		mux.Delete("/{userID}", app.deleteUser)
		mux.Put("/{userID}", app.updateUser)
	})

	return mux
}
