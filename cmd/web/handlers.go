package main

import "net/http"

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "signup.html", app.newTemplateData(r))
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Show signup page"))
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Post on sign up form"))
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Show login page"))
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Post on login form"))
}
