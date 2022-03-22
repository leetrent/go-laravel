package main

import (
	"errors"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/leetrent/celeritas"
)

const version = "1.0.0"

var cel celeritas.Celeritas

func main() {
	arg1, arg2, arg3, err := validateInput()
	if err != nil {
		exitGracefully(err)
	}

	setup()

	switch arg1 {
	case "help":
		showHelp()
	case "version":
		showVersion()
	case "make":
		if arg2 == "" {
			exitGracefully(errors.New("make requires a subcommand: [migration|model|handler]"))
		}
		err = doMake(arg2, arg3)
		if err != nil {
			exitGracefully(err)
		}
	default:
		log.Println(arg2, arg3)
	}
}

func validateInput() (string, string, string, error) {
	var arg1, arg2, arg3 string

	if len(os.Args) <= 1 {
		color.Red("Error: command required")
		showHelp()
		return "", "", "", errors.New("command required")
	}

	arg1 = os.Args[1]
	if len(os.Args) >= 3 {
		arg2 = os.Args[2]
	}
	if len(os.Args) >= 4 {
		arg3 = os.Args[3]
	}

	return arg1, arg2, arg3, nil
}

func showHelp() {
	color.Yellow(`Available commands:
	help		- show the help commands
	version		- print application version
	`)
}

func showVersion() {
	color.Yellow("Application version: " + version)
}

func exitGracefully(err error, msg ...string) {
	message := ""

	if len(msg) > 0 {
		message = msg[0]
	}

	if err != nil {
		color.Red("Error: %v\n", err)
	}

	if len(message) > 0 {
		color.Yellow(message)
	} else {
		color.Green("Finished!")
	}

	os.Exit(0)
}
