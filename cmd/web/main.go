package main

import (
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

// //Home is the home page handler
// func Home(w http.ResponseWriter, r *http.Request) {
// 	renderTemplate(w, "home.page.tmpl")
// }

// //About is the about page handler
// func About(w http.ResponseWriter, r *http.Request) {
// 	renderTemplate(w, "about.page.tmpl")
// }

// func renderTemplate(w http.ResponseWriter, tmpl string) {
// 	parsedTemplate, _ := template.ParseFiles("./templates/" + tmpl)
// 	err := parsedTemplate.Execute(w, nil)
// 	if err != nil {
// 		fmt.Println("Error parsing template: ", err)
// 		return
// 	}
// }

//main is the main application function
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()
	defer db.Neo4j.Close()

	fmt.Println(fmt.Sprintf("Starting application on port %s", portNumber))
	// fmt.Println(db.Neo4j.Target())

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
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

	// connect to databases
	log.Println("Connecting to database...")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=myApp1 user=USER password=PASSWORD")
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}
	log.Println("Connected to postgre SQL database!")

	db, err = driver.ConnectNeo4j("neo4j://localhost:7687", "DATABASE", "PASSWORD")
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}
	log.Println("Connected to Neo4j database!")

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
