package celeritas

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

type Celeritas struct {
	AppName  string
	Debug    bool
	Version  string
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	RootPath string
}

func (c *Celeritas) New(rootPath string) error {
	logSnippet := "\n[celeritas][New] =>"
	fmt.Printf("%s (rootPath): %s\n", logSnippet, rootPath)

	//////////////////////////////////////////////////////////
	// ASSIGN APPLICATION NAME
	//////////////////////////////////////////////////////////
	c.AppName = os.Getenv("APP_NAME")
	fmt.Printf("%s (os.Getenv(\"APP_NAME\")): %s\n", logSnippet, os.Getenv("APP_NAME"))
	fmt.Printf("%s (c.AppName): %s\n", logSnippet, c.AppName)

	//////////////////////////////////////////////////////////
	// ASSIGN APPLICATION VERSION
	//////////////////////////////////////////////////////////
	c.Version = version

	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middleware"},
	}

	//////////////////////////////////////////////////////////
	// Create application folders if they don't already exist
	//////////////////////////////////////////////////////////
	err := c.Init(pathConfig)
	if err != nil {
		return err
	}

	//////////////////////////////////////////////////////////
	// Create .env file if it doen't already exist
	//////////////////////////////////////////////////////////
	err = c.checkDotEnv(rootPath)
	if err != nil {
		return err
	}

	//////////////////////////////////////////////////////////
	// Read contents of .env file and create an
	// environment variable for each entry in .env file
	//////////////////////////////////////////////////////////
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	//////////////////////////////////////////////////////////
	// CREATE LOGGERS
	//////////////////////////////////////////////////////////
	infoLog, errorLog := c.startLoggers()
	c.InfoLog = infoLog
	c.ErrorLog = errorLog

	//////////////////////////////////////////////////////////
	// READ ENVIRONMENT VARIABLES AND ASSIGN VALUES TO
	// CORRESPONDING MEMBERS OF Celeritas struct
	//////////////////////////////////////////////////////////
	c.Debug, err = strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		c.ErrorLog.Println(err)
		return err
	}

	return nil
}

func (c *Celeritas) Init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		err := c.CreateDirIfNotExist(root + "/" + path)
		if err != nil {
			return err
		}

	}
	return nil
}

func (c *Celeritas) checkDotEnv(path string) error {
	err := c.CreateFileIfNotExists(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) startLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}
