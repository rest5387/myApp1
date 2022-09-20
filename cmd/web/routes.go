package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/rest5387/myApp1/internal/config"
	"github.com/rest5387/myApp1/internal/handlers"
)

func routes(app *config.AppConfig) http.Handler {

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	// Home page & card-wall ops.(Post, Update & Delete cards)
	mux.Get("/", handlers.Repo.Home)
	// User cards page
	mux.Route("/userid={userid}", func(mux chi.Router) {
		mux.Use(CardParamCtx)
		mux.Get("/", handlers.Repo.User)
	})
	// Log in/out page
	mux.Get("/login", handlers.Repo.Login)
	mux.Post("/login", handlers.Repo.PostLogin)
	mux.Get("/logout", handlers.Repo.Logout)
	// Sign up page
	mux.Get("/signup", handlers.Repo.SignUp)
	mux.Post("/signup", handlers.Repo.PostSignUp)

	// mux.Route("/CardAJAX", func(mux chi.Router) {
	// 	mux.Post("/", handlers.Repo.PostCardAJAX)
	// 	mux.Route("/offset={offset}", func(mux chi.Router) {
	// 		mux.Use(CardParamCtx)
	// 		mux.Get("/", handlers.Repo.GetCardAJAX)
	// 	})
	// 	mux.Route("/pid={pid}", func(mux chi.Router) {
	// 		mux.Use(CardParamCtx)
	// 		mux.Put("/", handlers.Repo.PutCardAJAX)
	// 		mux.Delete("/", handlers.Repo.DeleteCardAJAX)
	// 	})
	// 	mux.Route("/userid={userid}&offset={offset}", func(mux chi.Router) {
	// 		mux.Use(CardParamCtx)
	// 		mux.Get("/", handlers.Repo.GetCardAJAX)
	// 	})
	// })
	// mux.Route("/User/userid={userid}", func(mux chi.Router) {
	// 	mux.Use(CardParamCtx)
	// 	mux.Get("/", handlers.Repo.GetUser)
	// })
	// mux.Route("/Follow/userid={userid}", func(mux chi.Router) {
	// 	mux.Use(CardParamCtx)
	// 	mux.Post("/", handlers.Repo.PostFollow)
	// 	mux.Delete("/", handlers.Repo.DeleteFollow)
	// })

	// API routes
	mux.Route("/api", func(mux chi.Router) {
		mux.Route("/Card", func(mux chi.Router) {
			mux.Post("/", handlers.Repo.PostCardAJAX)
			mux.Route("/offset={offset}", func(mux chi.Router) {
				mux.Use(CardParamCtx)
				mux.Get("/", handlers.Repo.GetCardAJAX)
			})
			mux.Route("/pid={pid}", func(mux chi.Router) {
				mux.Use(CardParamCtx)
				mux.Put("/", handlers.Repo.PutCardAJAX)
				mux.Delete("/", handlers.Repo.DeleteCardAJAX)
			})
			mux.Route("/userid={userid}&offset={offset}", func(mux chi.Router) {
				mux.Use(CardParamCtx)
				mux.Get("/", handlers.Repo.GetCardAJAX)
			})
		})
		mux.Route("/User/userid={userid}", func(mux chi.Router) {
			mux.Use(CardParamCtx)
			mux.Get("/", handlers.Repo.GetUser)
		})
		mux.Route("/Follow/userid={userid}", func(mux chi.Router) {
			mux.Use(CardParamCtx)
			mux.Post("/", handlers.Repo.PostFollow)
			mux.Delete("/", handlers.Repo.DeleteFollow)
		})
	})

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	jsFileServer := http.FileServer(http.Dir("./js/"))
	mux.Handle("/js/*", http.StripPrefix("/js", jsFileServer))

	return mux
}
