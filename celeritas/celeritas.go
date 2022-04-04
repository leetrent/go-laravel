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
	"github.com/dgraph-io/badger/v3"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/leetrent/celeritas/cache"
	"github.com/leetrent/celeritas/mailer"
	"github.com/leetrent/celeritas/render"
	"github.com/leetrent/celeritas/session"
	"github.com/robfig/cron/v3"
)

const version = "1.0.0"

var myRedisCache *cache.RedisCache
var myBadgerCache *cache.BadgerCache
var redisPool *redis.Pool
var badgerConn *badger.DB

type Celeritas struct {
	AppName       string
	Debug         bool
	Version       string
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	RootPath      string
	Routes        *chi.Mux
	Render        *render.Render
	Session       *scs.SessionManager
	DB            Database
	JetViews      *jet.Set
	config        config
	EncryptionKey string
	Cache         cache.Cache
	Scheduler     *cron.Cron
	Mail          mailer.Mail
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	database    databaseConfig
	redis       redisConig
}

func (c *Celeritas) New(rootPath string) error {
	logSnippet := "\n[celeritas][celeritas.go][New()] =>"
	// fmt.Printf("%s (os.Getenv(\"SMTP_HOST\"): %s)\n", logSnippet, os.Getenv("SMTP_HOST"))
	// fmt.Printf("%s (rootPath)..: %s\n", logSnippet, rootPath)

	//////////////////////////////////////////////////////////
	// ASSIGN APPLICATION ROOT PATH
	//////////////////////////////////////////////////////////
	c.RootPath = rootPath
	//fmt.Printf("%s (c.RootPath): %s\n", logSnippet, c.RootPath)

	//////////////////////////////////////////////////////////
	// ASSIGN APPLICATION VERSION
	//////////////////////////////////////////////////////////
	c.Version = version

	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "mail", "data", "public", "tmp", "logs", "middleware"},
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
	//err = godotenv.Load(rootPath + "/.env")
	err = godotenv.Load(rootPath + "\\.env")
	if err != nil {
		return err
	}
	//fmt.Printf("%s (rootPath+\"\\.env\"): %s\n", logSnippet, rootPath+"\\.env")
	fmt.Printf("%s (os.Getenv(\"SMTP_HOST\"): %s\n", logSnippet, os.Getenv("SMTP_HOST"))

	//////////////////////////////////////////////////////////
	// CREATE LOGGERS
	//////////////////////////////////////////////////////////
	infoLog, errorLog := c.startLoggers()
	c.InfoLog = infoLog
	c.ErrorLog = errorLog

	//////////////////////////////////////////////////////////
	// CONNECT TO DATABASE
	//////////////////////////////////////////////////////////
	if os.Getenv("DATABASE_TYPE") != "" {
		db, err := c.OpenDB(os.Getenv("DATABASE_TYPE"), c.BuildDSN())
		if err != nil {
			c.ErrorLog.Println(err)
			os.Exit(1)
		}
		c.DB = Database{
			DataType: os.Getenv("DATABASE_TYPE"),
			Pool:     db,
		}
	}

	//////////////////////////////////////////////////////////
	// CONNECT TO REDIS CACHE
	//////////////////////////////////////////////////////////
	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		myRedisCache = c.createClientRedisCache()
		c.Cache = myRedisCache
		redisPool = myRedisCache.Conn
	}

	//////////////////////////////////////////////////////////
	// CONNECT TO BADGER CACHE
	//////////////////////////////////////////////////////////
	scheduler := cron.New()
	c.Scheduler = scheduler
	if os.Getenv("CACHE") == "badger" {
		myBadgerCache = c.createClientBadgerCache()
		c.Cache = myBadgerCache
		badgerConn = myBadgerCache.Conn

		_, err = c.Scheduler.AddFunc("@daily", func() {
			_ = myBadgerCache.Conn.RunValueLogGC(0.7)
		})
		if err != nil {
			return err
		}
	}

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
		database: databaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn:      c.BuildDSN(),
		},
		redis: redisConig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}

	//c.InfoLog.Printf("%s (c.config.port): %s\n", logSnippet, c.config.port)
	//c.InfoLog.Printf("%s (c.config.renderer): %s\n", logSnippet, c.config.renderer)

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

	switch c.config.sessionType {
	case "redis":
		httpSession.RedisPool = myRedisCache.Conn
	case "mysql", "postgres", "mariadb", "postgresql":
		httpSession.DBPool = c.DB.Pool
	}

	c.Session = httpSession.InitSession()

	//////////////////////////////////////////////////////////
	// READ ENCRYPTION KEY FROM .env FILE
	//////////////////////////////////////////////////////////
	c.EncryptionKey = os.Getenv("KEY")
	//c.EncryptionKey = "7zllP1TbvJv99l1xRJfHVtxff7ZfdX9d"
	// fmt.Println("")
	// fmt.Printf("c.EncryptionKey.......: '%s'", c.EncryptionKey)
	// fmt.Printf("\nlen(c.EncryptionKey): '%d'", len(c.EncryptionKey))
	// fmt.Println("")

	/////////////////////////////////////////////////////
	// ASSIGN AVAILABLE ROUTES
	//////////////////////////////////////////////////////////
	c.Routes = c.routes().(*chi.Mux)

	//////////////////////////////////////////////////////////
	// ASSIGN JET VIEWS
	//////////////////////////////////////////////////////////
	var views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		//jet.InDevelopmentMode(),
	)
	c.JetViews = views

	//logSnippet := "\n[celeritas][New] =>"
	// fmt.Println("")
	// fmt.Printf("%s (c.Debug): '%t'", logSnippet, c.Debug)
	// fmt.Println("")

	if c.Debug {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
			jet.InDevelopmentMode(),
		)
		c.JetViews = views
	} else {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		)
		c.JetViews = views
	}

	//////////////////////////////////////////////////////////
	// ASSIGN TEMPLATE RENDERER
	//////////////////////////////////////////////////////////
	c.createRenderer()

	//////////////////////////////////////////////////////////
	// MAIL SERVICE
	//////////////////////////////////////////////////////////
	c.Mail = c.createMailer()
	go c.Mail.ListenForMail()

	return nil
}

func (c *Celeritas) Init(p initPaths) error {
	// logSnippet := "\n[celeritas][Init] =>"
	// fmt.Printf("%s (p.rootPath)..: %s\n", logSnippet, p.rootPath)

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
	// logSnippet := "\n[celeritas][checkDotEnv] =>"
	// fmt.Printf("%s (path)..: %s\n", logSnippet, path)

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

	/////////////////////////////////////////////
	// CLOSE DATABASE WHEN APPLICTION SHUTS DOWN
	/////////////////////////////////////////////
	if c.DB.Pool != nil {
		defer c.DB.Pool.Close()
	}

	////////////////////////////////////////////////
	// CLOSE REDIS CACHE WHEN APPLICTION SHUTS DOWN
	////////////////////////////////////////////////
	if redisPool != nil {
		defer redisPool.Close()
	}

	////////////////////////////////////////////////
	// CLOSE BADGER CACHE WHEN APPLICTION SHUTS DOWN
	////////////////////////////////////////////////
	if badgerConn != nil {
		defer badgerConn.Close()
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
		Session:  c.Session,
	}
	c.Render = &myRenderer
}

func (c *Celeritas) createMailer() mailer.Mail {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	m := mailer.Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Templates:   c.RootPath + "/mail",
		Host:        os.Getenv("SMTP_HOST"),
		Port:        port,
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		Encryption:  os.Getenv("SMTP_ENCRYPTION"),
		FromName:    os.Getenv("FROM_NAME"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
		Jobs:        make(chan mailer.Message, 20),
		Results:     make(chan mailer.Result, 20),
		API:         os.Getenv("MAILER_API"),
		APIKey:      os.Getenv("MAILER_KEY"),
		APIUrl:      os.Getenv("MAILER_URL"),
	}

	logSnippet := "\n[celerita][celeritas.go][createMailer()] =>"
	fmt.Printf("%s (os.Getenv(\"SMTP_HOST\"): %s\n", logSnippet, os.Getenv("SMTP_HOST"))
	fmt.Printf("%s (m.Host)...............: %s\n", logSnippet, m.Host)
	return m
}

func (c *Celeritas) BuildDSN() string {
	var dsn string

	//databaseType := os.Getenv("DATABASE_TYPE")
	//c.InfoLog.Printf("[celeritas][BuildDSN]: (databaseType): '%s';", databaseType)

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"))

		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}

	default:
	}

	//c.InfoLog.Printf("[celeritas][BuildDSN]: (dsn): %s", dsn)
	return dsn
}

func (c *Celeritas) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				c.config.redis.host,
				redis.DialPassword(c.config.redis.password))
		},

		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
	}
}

func (c *Celeritas) createBadgerConn() *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(c.RootPath + "/tmp/badger"))
	if err != nil {
		return nil
	}
	return db
}

func (c *Celeritas) createClientRedisCache() *cache.RedisCache {
	cacheClient := cache.RedisCache{
		Conn:   c.createRedisPool(),
		Prefix: c.config.redis.prefix,
	}
	return &cacheClient
}

func (c *Celeritas) createClientBadgerCache() *cache.BadgerCache {
	cacheClient := cache.BadgerCache{
		Conn: c.createBadgerConn(),
	}
	return &cacheClient
}
