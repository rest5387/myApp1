package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/justinas/nosurf"
)

// NoSurf adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	csrfHandler.ExemptFunc(func(r *http.Request) bool {
		if r.Method == "PUT" || r.Method == "DELETE" {
			return true
		}
		return false
	})

	csrfHandler.ExemptGlob("/api/Follow/*")

	return csrfHandler
}

// SessionLoad loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
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
