package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rest5387/myApp1/internal/helpers"

	"github.com/rest5387/myApp1/internal/models"

	"github.com/alexedwards/scs/v2"

	"github.com/rest5387/myApp1/internal/config"
	"github.com/rest5387/myApp1/internal/render"

	"github.com/rest5387/myApp1/internal/driver"
	"github.com/rest5387/myApp1/internal/handlers"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

//main is the main application function
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()
	defer db.Neo4j.Close()

	fmt.Println(fmt.Sprintf("Starting application on port %s", portNumber))

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	var (
		neo4j_host     string
		neo4j_username string
		neo4j_password string
		sql_dsn        string
		redis_host     string
		redis_password string
	)

	// what I am going to put in the session
	gob.Register(models.Reservation{})
	// change this to truw when in production
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// Read DBs setting from file.
	file, err := os.Open("DB_setting.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		str := scanner.Text()
		switch str {
		case "neo4j":
			{
				scanner.Scan()
				str = scanner.Text()
				fmt.Sscanf(str, "host:%s", &neo4j_host)
				scanner.Scan()
				str = scanner.Text()
				fmt.Sscanf(str, "username:%s", &neo4j_username)
				scanner.Scan()
				str = scanner.Text()
				fmt.Sscanf(str, "password:%s", &neo4j_password)
			}
		case "postgresql":
			{
				scanner.Scan()
				sql_dsn = scanner.Text()
			}
		case "redis":
			{
				scanner.Scan()
				str = scanner.Text()
				fmt.Sscanf(str, "host:%s", &redis_host)
				scanner.Scan()
				str = scanner.Text()
				fmt.Sscanf(str, "password:%s", &redis_password)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// connect to databases
	log.Println("Connecting to database...")
	db, err := driver.ConnectSQL(sql_dsn)
	if err != nil {
		log.Fatal("Cannot connect to postgre SQL database! Dying...")
	}
	log.Println("Connected to postgre SQL database!")

	db, err = driver.ConnectNeo4j(neo4j_host, neo4j_username, neo4j_password)
	if err != nil {
		log.Fatal("Cannot connect to Neo4j database! Dying...")
	}
	log.Println("Connected to Neo4j database!")

	db, err = driver.ConnectRedis(redis_host, redis_password, 0)
	if err != nil {
		log.Fatal("Cannot connect to Reids database! Dying...")
	}
	log.Println("Connected to Redis database!")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache", err)
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewTemplates(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
