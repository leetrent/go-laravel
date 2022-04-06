package main

import (
	"fmt"
	"myapp/data"
	"net/http"
	"strconv"

	"github.com/leetrent/celeritas/mailer"

	"github.com/go-chi/chi/v5"
)

func (a *application) routes() *chi.Mux {
	//////////////////////////////////////////
	// MIDDLEWARE MUST COME BEFORE ANY ROUTES
	//////////////////////////////////////////

	//////////////////////////////////////////
	// ADD ROUTES HERE
	//////////////////////////////////////////
	a.get("/", a.Handlers.Home)
	a.App.Routes.Get("/go-page", a.Handlers.GoPage)
	a.App.Routes.Get("/jet-page", a.Handlers.JetPage)
	a.App.Routes.Get("/sessions", a.Handlers.SessionTest)
	a.App.Routes.Get("/users/login", a.Handlers.UserLogin)
	a.post("/users/login", a.Handlers.PostUserLogin)
	a.App.Routes.Get("/users/logout", a.Handlers.UserLogout)

	a.App.Routes.Get("/form", a.Handlers.Form)
	a.App.Routes.Post("/form", a.Handlers.PostForm)

	///////////////////////////////////////////////
	// TEST JSON, XML, File Download functionality
	///////////////////////////////////////////////
	a.get("/json", a.Handlers.JSON)
	a.get("/xml", a.Handlers.XML)
	a.get("/download-file", a.Handlers.DownloadFile)

	///////////////////////////////////////////////
	// TEST ENCRYPTION, DECRYPTION FUNCTIONALITY
	///////////////////////////////////////////////
	a.get("/crypto", a.Handlers.TestCrypto)

	///////////////////////////////////////////////
	// TEST REDIS CACHE FUNCTIONALITY
	///////////////////////////////////////////////
	a.get("/cache-test", a.Handlers.ShowCachePage)
	a.post("/api/save-in-cache", a.Handlers.ShowCachePage)
	a.post("/api/save-in-cache", a.Handlers.SaveInCache)
	a.post("/api/get-from-cache", a.Handlers.GetFromCache)
	a.post("/api/delete-from-cache", a.Handlers.DeleteFromCache)
	a.post("/api/empty-cache", a.Handlers.EmptyCache)

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
	// TEST SMTP MAIL SERVICE
	///////////////////////////////////////////////////////////////////////////////
	a.get("/test-mail", func(w http.ResponseWriter, r *http.Request) {
		// msg := mailer.Message{
		// 	From:        "test@example.com",
		// 	To:          "you@there.com",
		// 	Subject:     "Test Subject - sent using channel",
		// 	Template:    "test",
		// 	Attachments: nil,
		// 	Data:        nil,
		// }
		msg := mailer.Message{
			From:        "mailgun@sandbox2cbe917537114f7d8a25a031c42732e0.mailgun.org",
			To:          "lee.trent@icloud.com",
			Subject:     "Test Subject - sent using an API via channel",
			Template:    "test",
			Attachments: nil,
			Data:        nil,
		}

		///////////////////////////////////////////////////////////////////////////////
		// SEND MAIL USING CHANNEL
		///////////////////////////////////////////////////////////////////////////////
		a.App.Mail.Jobs <- msg
		res := <-a.App.Mail.Results
		if res.Error != nil {
			a.App.ErrorLog.Println(res.Error)
		}
		///////////////////////////////////////////////////////////////////////////////

		///////////////////////////////////////////////////////////////////////////////
		// SEND MAIL NOT USING CHANNEL (DIRECT CALL)
		///////////////////////////////////////////////////////////////////////////////
		// err := a.App.Mail.SendSMTPMessage(msg)
		// if err != nil {
		// 	a.App.ErrorLog.Println(err)
		// 	return
		// }
		///////////////////////////////////////////////////////////////////////////////

		fmt.Fprint(w, "Sent mail!")
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

		/////////////////////////////////////////////////////////////////////////////////////////////
		// TEST VALIDATOR
		/////////////////////////////////////////////////////////////////////////////////////////////
		validator := a.App.Validator(nil)
		//validator.Check(len(u.LastName) > 20, "last_name", "Last name must be 20 characters or more")

		u.LastName = ""
		u.Validate(validator)
		if !validator.Valid() {
			fmt.Fprint(w, "failed validation")
			return
		}
		/////////////////////////////////////////////////////////////////////////////////////////////

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
