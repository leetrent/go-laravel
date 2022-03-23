package main

import (
	"fmt"
	"log"
	"time"
)

func doAuth() error {
	////////////////////////////////////////////////////////////////////////
	// CREATE MIGRATIONS
	////////////////////////////////////////////////////////////////////////
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

	err = copyDataToFile([]byte("drop table if exists users cascade"), downFile)
	if err != nil {
		exitGracefully(err)
	}

	////////////////////////////////////////////////////////////////////////
	// RUN MIGRATIONS
	////////////////////////////////////////////////////////////////////////
	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}

	////////////////////////////////////////////////////////////////////////
	// COPY FILES
	////////////////////////////////////////////////////////////////////////
	return nil
}
