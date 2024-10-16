package main

import (
	"net/http"

	"gochat.ayonchakroborty.net/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "about.html", app.newTemplateData(r))
}

type userSignupForm struct {
	UserName            string `form:"username"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, r, http.StatusOK, "signup.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil{
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var form userSignupForm

	err = app.formDecoder.Decode(&form, r.PostForm)
	if err != nil{
		app.clientError(w, http.StatusBadGateway)
		return
	}

	form.CheckField(validator.NotBlank(form.UserName), "username", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid(){
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Show login page"))
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Post on login form"))
}
