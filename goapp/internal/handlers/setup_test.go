package handlers

import (
	"bufio"
	"context"
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
	"github.com/rest5387/myApp1/internal/config"
	"github.com/rest5387/myApp1/internal/driver"
	"github.com/rest5387/myApp1/internal/helpers"
	"github.com/rest5387/myApp1/internal/models"
	"github.com/rest5387/myApp1/internal/render"
)

var app config.AppConfig
var session *scs.SessionManager
var pathToTemplates = "./../../templates" // run tests in dir handlers, not in bookings root dir
var functions = template.FuncMap{}

func getRoutes() http.Handler {
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

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
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

	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache", err)
	}

	app.TemplateCache = tc
	app.UseCache = true

	repo := NewRepo(&app, db)
	NewHandlers(repo)
	render.NewTemplates(&app)
	helpers.NewHelpers(&app)

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	// mux.Use(NoSurf) // Close NoSurf middleware for post request handler test (CSRFToken check)
	mux.Use(SessionLoad)

	// Home page & card-wall ops.(Post, Update & Delete cards)
	mux.Get("/", Repo.Home)
	// User cards page
	mux.Route("/userid={userid}", func(mux chi.Router) {
		mux.Use(CardParamCtx)
		mux.Get("/", Repo.User)
	})
	// Log in/out page
	mux.Get("/login", Repo.Login)
	mux.Post("/login", Repo.PostLogin)
	mux.Get("/logout", Repo.Logout)
	// Sign up page
	mux.Get("/signup", Repo.SignUp)
	mux.Post("/signup", Repo.PostSignUp)

	// API routes
	mux.Route("/api", func(mux chi.Router) {
		mux.Route("/Card", func(mux chi.Router) {
			mux.Post("/", Repo.PostCardAJAX)
			mux.Route("/offset={offset}", func(mux chi.Router) {
				mux.Use(CardParamCtx)
				mux.Get("/", Repo.GetCardAJAX)
			})
			mux.Route("/pid={pid}", func(mux chi.Router) {
				mux.Use(CardParamCtx)
				mux.Put("/", Repo.PutCardAJAX)
				mux.Delete("/", Repo.DeleteCardAJAX)
			})
			mux.Route("/userid={userid}&offset={offset}", func(mux chi.Router) {
				mux.Use(CardParamCtx)
				mux.Get("/", Repo.GetCardAJAX)
			})
		})
		mux.Route("/User/userid={userid}", func(mux chi.Router) {
			mux.Use(CardParamCtx)
			mux.Get("/", Repo.GetUser)
		})
		mux.Route("/Follow/userid={userid}", func(mux chi.Router) {
			mux.Use(CardParamCtx)
			mux.Post("/", Repo.PostFollow)
			mux.Delete("/", Repo.DeleteFollow)
		})
	})

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	jsFileServer := http.FileServer(http.Dir("./js/"))
	mux.Handle("/js/*", http.StripPrefix("/js", jsFileServer))

	return mux

}

// NoSurf adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

// SessionLoad loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

//CreateTestTemplateCache creates a template cache as a map
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))

	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		// fmt.Println("Page is currently", page)
		name := filepath.Base(page)
		// fmt.Println("Page is currently", name)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)

		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}
		myCache[name] = ts
	}
	return myCache, nil
}

// Get parameters from request URL
func CardParamCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Compare(r.Method, "GET") == 0 || strings.Compare(r.Method, "POST") == 0 {
			userIDStr := chi.URLParam(r, "userid")
			offsetStr := chi.URLParam(r, "offset")
			userID, _ := strconv.Atoi(userIDStr)
			offset, _ := strconv.Atoi(offsetStr)
			ctx := context.WithValue(r.Context(), "userid", userID)
			ctx = context.WithValue(ctx, "offset", offset)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		if strings.Compare(r.Method, "PUT") == 0 || strings.Compare(r.Method, "DELETE") == 0 {
			pidStr := chi.URLParam(r, "pid")
			pid, _ := strconv.Atoi(pidStr)
			userIDStr := chi.URLParam(r, "userid")
			userID, _ := strconv.Atoi(userIDStr)
			ctx := context.WithValue(r.Context(), "pid", pid)
			ctx = context.WithValue(ctx, "userid", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
