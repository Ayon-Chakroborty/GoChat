package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))

	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /user/signup", app.userSignup)
	mux.HandleFunc("POST /user/signup", app.userSignupPost)
	mux.HandleFunc("GET /user/login", app.userLogin)
	mux.HandleFunc("POST /user/login", app.userLoginPost)

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return standard.Then(mux)
}
