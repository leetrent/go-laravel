package main

import (
	"fmt"
	"log"
	"myapp/data"
	"myapp/handlers"
	"myapp/middleware"
	"os"

	"github.com/leetrent/celeritas"
)

func initApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	logSnippet := "[myapp][init-celeritas] =>"
	fmt.Println("")
	fmt.Printf("%s os.Getwd(): %s", logSnippet, path)
	fmt.Println("")

	// init celeritas
	cel := &celeritas.Celeritas{}
	err = cel.New(path)
	if err != nil {
		log.Fatal(err)
	}

	cel.AppName = "myapp"

	myMiddleware := &middleware.Middleware{
		App: cel,
	}

	// cel.InfoLog.Println("AppName is set to:", cel.AppName)
	// cel.InfoLog.Println("Version is set to:", cel.Version)
	// cel.InfoLog.Println("Debug is set to..:", cel.Debug)

	myHandlers := &handlers.Handlers{
		App: cel,
	}

	app := &application{
		App:        cel,
		Handlers:   myHandlers,
		Middleware: myMiddleware,
	}

	app.App.Routes = app.routes()

	app.Models = data.New(app.App.DB.Pool)
	myHandlers.Models = app.Models
	app.Middleware.Models = app.Models

	return app
}
