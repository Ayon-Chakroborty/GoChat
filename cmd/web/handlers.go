package main

import (
	"errors"
	"net/http"
	"strings"

	"gochat.ayonchakroborty.net/internal/models"
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
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var form userSignupForm

	err = app.formDecoder.Decode(&form, r.PostForm)
	if err != nil {
		app.clientError(w, http.StatusBadGateway)
		return
	}

	form.CheckField(validator.NotBlank(form.UserName), "username", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	err = app.userModel.Insert(form.UserName, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Account created successfully!")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.html", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := userLoginForm{}

	err = app.formDecoder.Decode(&form, r.PostForm)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	id, err := app.userModel.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or Password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
			return
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	username, err := app.userModel.GetUserField("username", form.Email)
	if err != nil {
		app.serverError(w, r, err)
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	app.sessionManager.Put(r.Context(), "email", form.Email)
	app.sessionManager.Put(r.Context(), "username", username)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Remove(r.Context(), "email")
	app.sessionManager.Remove(r.Context(), "username")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) chat(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "chat.html", app.newTemplateData(r))
}

func (app *application) userAccount(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	username, err := app.userModel.GetUserField("username", data.Email)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	form := userSignupForm{
		UserName: username,
		Email:    data.Email,
	}
	data.Form = form
	app.render(w, r, http.StatusOK, "account.html", data)
}

func (app *application) userAccountPost(w http.ResponseWriter, r *http.Request) {
	email := app.sessionManager.GetString(r.Context(), "email")

	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := userSignupForm{}

	if err := app.formDecoder.Decode(&form, r.PostForm); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if validator.NotBlank(form.Email) {
		form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email")
	}
	if validator.NotBlank(form.Password) {
		form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
	}
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "account.html", data)
		return
	}

	formValues := map[string]string{
		"email":    form.Email,
		"username": form.UserName,
		"password": form.Password,
	}

	for field, val := range formValues {
		if !validator.NotBlank(val) {
			continue
		}
		err := app.userModel.UpdateField(field, val, email)
		if err != nil {
			if errors.Is(err, models.ErrDuplicateEmail) {
				form.AddFieldError("email", "Email address is already in use")

				data := app.newTemplateData(r)
				data.Form = form
				app.render(w, r, http.StatusUnprocessableEntity, "account.html", data)
			} else {
				app.serverError(w, r, err)
			}

			return
		}

		if strings.Compare(field, "email") == 0 {
			email = val
			app.sessionManager.Put(r.Context(), "email", val)
		} else if strings.Compare(field, "username") == 0{
			app.sessionManager.Put(r.Context(), "username", val)
		}
	}

	app.sessionManager.Put(r.Context(), "flash", "Account info changed successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)

}
