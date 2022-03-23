package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
)

func doAuth() error {
	//////////////////////////////////////////////////////////////////////////////////////////////////
	// CREATE MIGRATIONS
	//////////////////////////////////////////////////////////////////////////////////////////////////
	dbType := cel.DB.DataType
	fileName := fmt.Sprintf("%d_create_auth_tables", time.Now().UnixMicro())
	upFile := cel.RootPath + "/migrations/" + fileName + ".up.sql"
	downFile := cel.RootPath + "/migrations/" + fileName + ".down.sql"

	log.Println("dbType..:", dbType)
	log.Println("fileName:", fileName)
	log.Println("upFile..:", upFile)
	log.Println("downFile:", downFile)

	err := copyFileFromTemplate("templates/migrations/auth_tables."+dbType+".sql", upFile)
	if err != nil {
		exitGracefully(err)
	}

	err = copyDataToFile([]byte("drop table if exists users cascade; drop table if exists tokens cascade; drop table if exists remember_tokens;"), downFile)
	if err != nil {
		exitGracefully(err)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////////
	// RUN MIGRATIONS
	//////////////////////////////////////////////////////////////////////////////////////////////////
	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////////
	// COPY OVER FILES
	//////////////////////////////////////////////////////////////////////////////////////////////////
	err = copyFileFromTemplate("templates/data/user.go.txt", cel.RootPath+"/data/user.go")
	if err != nil {
		exitGracefully(err)
	}
	err = copyFileFromTemplate("templates/data/token.go.txt", cel.RootPath+"/data/token.go")
	if err != nil {
		exitGracefully(err)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////////
	// COPY OVER MIDDLEWARE
	//////////////////////////////////////////////////////////////////////////////////////////////////
	err = copyFileFromTemplate("templates/middleware/auth.go.txt", cel.RootPath+"/middleware/auth.go")
	if err != nil {
		exitGracefully(err)
	}
	err = copyFileFromTemplate("templates/middleware/auth-token.go.txt", cel.RootPath+"/middleware/auth-token.go")
	if err != nil {
		exitGracefully(err)
	}

	color.Yellow("  - users, tokens and remember_tokens migrations created and executed")
	color.Yellow("  - user and token models created")
	color.Yellow("  - auth middleware created")
	color.Yellow("")
	color.Yellow("Don't forget to add user and token models in data/models.go.")
	color.Yellow("And don't forget to add appropriate middleware to your routes.")

	return nil
}
