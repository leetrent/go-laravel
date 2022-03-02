package celeritas

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/leetrent/celeritas/render"
	"github.com/leetrent/celeritas/session"
)

const version = "1.0.0"

type Celeritas struct {
	AppName  string
	Debug    bool
	Version  string
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	RootPath string
	Routes   *chi.Mux
	Render   *render.Render
	Session  *scs.SessionManager
	JetViews *jet.Set
	config   config
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
}

func (c *Celeritas) New(rootPath string) error {
	logSnippet := "\n[celeritas][New] =>"
	fmt.Printf("%s (rootPath): %s\n", logSnippet, rootPath)

	//////////////////////////////////////////////////////////
	// ASSIGN APPLICATION ROOT PATH
	//////////////////////////////////////////////////////////
	c.RootPath = rootPath

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
	// ASSIGN APPLICATION NAME
	//////////////////////////////////////////////////////////
	c.AppName = os.Getenv("APP_NAME")

	//////////////////////////////////////////////////////////
	// READ ENVIRONMENT VARIABLES AND ASSIGN VALUES TO
	// CORRESPONDING MEMBERS OF Celeritas struct
	//////////////////////////////////////////////////////////
	c.Debug, err = strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		c.ErrorLog.Println(err)
		return err
	}

	//////////////////////////////////////////////////////////
	// ASSIGN CONFIGURATION FOR CELERITAS
	//////////////////////////////////////////////////////////
	c.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSISTS"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
	}

	c.InfoLog.Printf("%s (c.config.port): %s\n", logSnippet, c.config.port)
	c.InfoLog.Printf("%s (c.config.renderer): %s\n", logSnippet, c.config.renderer)

	//////////////////////////////////////////////////////////
	// CREATE HTTP SESSION
	//////////////////////////////////////////////////////////
	httpSession := session.Session{
		CookieLifetime: c.config.cookie.lifetime,
		CookiePersist:  c.config.cookie.persist,
		CookieName:     c.config.cookie.name,
		SessionType:    c.config.sessionType,
		CookieDomain:   c.config.cookie.domain,
	}
	c.Session = httpSession.InitSession()

	/////////////////////////////////////////////////////
	// ASSIGN AVAILABLE ROUTES
	//////////////////////////////////////////////////////////
	c.Routes = c.routes().(*chi.Mux)

	//////////////////////////////////////////////////////////
	// ASSIGN JET VIEWS
	//////////////////////////////////////////////////////////
	var views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		jet.InDevelopmentMode(),
	)
	c.JetViews = views

	//////////////////////////////////////////////////////////
	// ASSIGN TEMPLATE RENDERER
	//////////////////////////////////////////////////////////
	c.createRenderer()

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

// ListentAndServe starts the web server
func (c *Celeritas) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		ErrorLog:     c.ErrorLog,
		Handler:      c.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}
	c.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))
	err := srv.ListenAndServe()
	c.ErrorLog.Fatal(err)
}

func (c *Celeritas) createRenderer() {
	myRenderer := render.Render{
		Renderer: c.config.renderer,
		RootPath: c.RootPath,
		Port:     c.config.port,
		JetViews: c.JetViews,
	}
	c.Render = &myRenderer
}
