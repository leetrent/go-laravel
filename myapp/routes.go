package main

import (
	"fmt"
	"myapp/data"
	"net/http"
	"strconv"

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
	a.App.Routes.Get("/sessions", a.Handlers.SessionTest)

	//////////////////////////////////////////
	// TEST DATABASE
	//////////////////////////////////////////
	a.App.Routes.Get("/test-database", func(w http.ResponseWriter, r *http.Request) {
		query := "SELECT id, first_name FROM users WHERE id = 1"
		row := a.App.DB.Pool.QueryRowContext(r.Context(), query)

		var id int
		var name string
		err := row.Scan(&id, &name)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		fmt.Fprintf(w, "id: '%d', name='%s'", id, name)
	})

	///////////////////////////////////////////////////////////////////////////////
	// TEST CREATE USER
	///////////////////////////////////////////////////////////////////////////////
	a.App.Routes.Get("/create-user", func(w http.ResponseWriter, r *http.Request) {
		u := data.User{
			FirstName: "Penelope",
			LastName:  "Trent",
			Email:     "penny@penny.com",
			Active:    1,
			Password:  "password",
		}
		id, err := a.Models.Users.Insert(u)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}
		fmt.Fprintf(w, "id: '%d', name: '%s'", id, u.FirstName)
	})

	///////////////////////////////////////////////////////////////////////////////
	// TEST GET ALL USERS
	///////////////////////////////////////////////////////////////////////////////
	a.App.Routes.Get("/get-all-users", func(w http.ResponseWriter, r *http.Request) {
		users, err := a.Models.Users.GetAll()
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		for _, x := range users {
			fmt.Fprint(w, x.LastName)
		}
	})

	///////////////////////////////////////////////////////////////////////////////
	// TEST GET USER BY ID
	///////////////////////////////////////////////////////////////////////////////
	a.App.Routes.Get("/get-user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		u, err := a.Models.Users.GetByID(id)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		fmt.Fprintf(w, "%s | %s | %s", u.FirstName, u.LastName, u.Email)
	})

	///////////////////////////////////////////////////////////////////////////////
	// TEST UPDATE USER BY ID
	///////////////////////////////////////////////////////////////////////////////
	a.App.Routes.Get("/update-user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		u, err := a.Models.Users.GetByID(id)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		u.LastName = a.App.RandomString(10)
		err = u.Update(*u)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		fmt.Fprintf(w, "%s | %s | %s", u.FirstName, u.LastName, u.Email)
	})

	//////////////////////////////////////////
	// STATIC ROUTES HERE
	//////////////////////////////////////////
	fileServer := http.FileServer(http.Dir("./public"))
	a.App.Routes.Handle("/public/*", http.StripPrefix("/public", fileServer))

	return a.App.Routes
}
