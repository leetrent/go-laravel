package main

import (
	"log"
	"myapp/handlers"
	"os"

	"github.com/leetrent/celeritas"
)

func initApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// init celeritas
	cel := &celeritas.Celeritas{}
	err = cel.New(path)
	if err != nil {
		log.Fatal(err)
	}

	// cel.InfoLog.Println("AppName is set to:", cel.AppName)
	// cel.InfoLog.Println("Version is set to:", cel.Version)
	// cel.InfoLog.Println("Debug is set to..:", cel.Debug)

	myHandlers := &handlers.Handlers{
		App: cel,
	}

	app := &application{
		App:      cel,
		Handlers: myHandlers,
	}

	app.App.Routes = app.routes()

	return app
}
