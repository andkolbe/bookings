package main

import (
	"net/http"

	"github.com/andkolbe/bookings/pkg/config"
	"github.com/andkolbe/bookings/pkg/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)


func routes(app *config.AppConfig) http.Handler {
	
	mux := chi.NewRouter()

	// middleware allows you process a request as it comes into your web app and perform some action on it
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)

	// create a file server - a place to get static files from
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("static/*", http.StripPrefix("/static", fileServer))

	return mux
}