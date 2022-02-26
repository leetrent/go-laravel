package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (a *application) routes() *chi.Mux {
	//////////////////////////////////////////
	// MIDDLEWARE MUST COME BEFORE ANY ROUTES
	//////////////////////////////////////////

	//////////////////////////////////////////
	// ADD ROUTES HERE
	//////////////////////////////////////////
	a.App.Routes.Get("/", a.Handlers.Home)
	a.App.Routes.Get("/go-page", a.Handlers.GoPage)
	a.App.Routes.Get("/jet-page", a.Handlers.JetPage)
	//////////////////////////////////////////
	// STATIC ROUTES HERE
	//////////////////////////////////////////
	fileServer := http.FileServer(http.Dir("./public"))
	a.App.Routes.Handle("/public/*", http.StripPrefix("/public", fileServer))

	return a.App.Routes
}
