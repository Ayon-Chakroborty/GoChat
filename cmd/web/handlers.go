package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"gochat.ayonchakroborty.net/internal/models"
	"gochat.ayonchakroborty.net/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	chatrooms, err := app.chatroomModel.GetAllChats(data.Email)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	publicChatrooms := []*models.Chatroom{}
	privateChatrooms := []*models.Chatroom{}

	for _, cr := range chatrooms {
		if cr.Private {
			privateChatrooms = append(privateChatrooms, cr)
		} else {
			publicChatrooms = append(publicChatrooms, cr)
		}
	}

	data.PublicChatrooms = publicChatrooms
	data.PrivateChatrooms = privateChatrooms

	for _, room := range data.PublicChatrooms {
		names, err := app.chatroomModel.GetUsersInChatroom(room.Name, room.Private)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		formattedString := strings.Join(names, ", ")
		room.AllUsers = formattedString
	}

	for _, room := range data.PrivateChatrooms {
		names, err := app.chatroomModel.GetUsersInChatroom(room.Name, room.Private)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		formattedString := strings.Join(names, ", ")
		room.AllUsers = formattedString
	}

	app.render(w, r, http.StatusOK, "home.html", data)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
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

	err = app.chatroomModel.Insert("general", form.Email, false)
	if err != nil {
		app.serverError(w, r, err)
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
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	app.sessionManager.Put(r.Context(), "email", form.Email)
	app.sessionManager.Put(r.Context(), "username", username)
	app.sessionManager.Put(r.Context(), "chatroom", "general")

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
	data := app.newTemplateData(r)
	chats, err := app.chatModel.Get(data.Chatroom)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	for _, chat := range chats {
		chat.Created = chat.Created.In(loc)
	}

	data.Chats = chats
	app.render(w, r, http.StatusOK, "chat.html", data)
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
		} else if strings.Compare(field, "username") == 0 {
			app.sessionManager.Put(r.Context(), "username", val)
		}
	}

	app.sessionManager.Put(r.Context(), "flash", "Account info changed successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type chatRoomForm struct {
	Chatroom            string `form:"chatroom"`
	validator.Validator `form:"-"`
}

func (app *application) chatRoom(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	app.sessionManager.Put(r.Context(), "chatroom", name)

	http.Redirect(w, r, "/chat", http.StatusSeeOther)
}

func (app *application) chatRoomPost(w http.ResponseWriter, r *http.Request) {
	email := app.sessionManager.GetString(r.Context(), "email")

	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := chatRoomForm{}

	if err := app.formDecoder.Decode(&form, r.PostForm); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if !validator.NotBlank(form.Chatroom) {
		http.Redirect(w, r, "/chat", http.StatusSeeOther)
		return
	}

	cr := form.Chatroom
	private := false
	if validator.Matches(form.Chatroom, validator.EmailRX) {
		if strings.Compare(form.Chatroom, email) == 0 {
			http.Redirect(w, r, "/chat", http.StatusSeeOther)
			return
		}

		exists, err := app.userModel.EmailExists(form.Chatroom)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		if !exists {
			flash := fmt.Sprintf("User '%s' does not exist", form.Chatroom)
			app.sessionManager.Put(r.Context(), "flash", flash)
			http.Redirect(w, r, "/chat", http.StatusSeeOther)
			return
		}

		sorted := []string{email, form.Chatroom}
		sort.Strings(sorted)
		cr = "Private chatroom for " + sorted[0] + " and " + sorted[1]
		private = true
	}

	_, err := app.chatroomModel.Get(cr, email, private)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			private := false
			// if the chatroom is a user email then insert 2 times for each user
			if validator.Matches(form.Chatroom, validator.EmailRX) {
				private = true
				err := app.chatroomModel.Insert(cr, form.Chatroom, private)
				if err != nil {
					log.Print("Error while inserting new chat room", err)
					http.Redirect(w, r, "/", http.StatusSeeOther)
				}
			}
			// public chat room
			err := app.chatroomModel.Insert(cr, email, private)
			if err != nil {
				log.Print("Error while inserting new chat room", err)
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
		} else {
			app.serverError(w, r, err)
			return
		}
	}

	app.sessionManager.Put(r.Context(), "chatroom", cr)

	http.Redirect(w, r, "/chat", http.StatusSeeOther)
}

func (app *application) userDeletePost(w http.ResponseWriter, r *http.Request) {
	email := app.sessionManager.GetString(r.Context(), "email")
	log.Println("here in userDeletePost")
	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	// if err := app.chatModel.DeleteUser(email); err != nil {
	// 	app.serverError(w, r, err)
	// 	return
	// }

	// if err := app.chatroomModel.DeleteUser(email); err != nil{
	// 	app.serverError(w, r, err)
	// 	return
	// }

	if err := app.userModel.DeleteUser(email); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Remove(r.Context(), "email")
	app.sessionManager.Remove(r.Context(), "username")
	app.sessionManager.Put(r.Context(), "flash", "Your account has been deleted successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type searchForm struct {
	Search string `form:"search"`
	validator.Validator
}

func (app *application) chatSearch(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusSeeOther, "search.html", app.newTemplateData(r))
}

func (app *application) chatSearchPost(w http.ResponseWriter, r *http.Request) {
	email := app.sessionManager.GetString(r.Context(), "email")

	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusUnprocessableEntity)
		return
	}

	form := searchForm{}
	if err := app.formDecoder.Decode(&form, r.PostForm); err != nil {
		app.clientError(w, http.StatusUnprocessableEntity)
		return
	}
	form.Search = strings.TrimSpace(form.Search)

	if !validator.NotBlank(form.Search) {
		http.Redirect(w, r, "/chat/search", http.StatusSeeOther)
		return
	}

	data := app.newTemplateData(r)
	chatrooms := []*models.Chatroom{}
	var err error

	if validator.Matches(form.Search, validator.EmailRX) {
		chatrooms, err = app.chatroomModel.SearchUser(email, form.Search)
	} else {
		chatrooms, err = app.chatroomModel.GetSearchedChat(email, form.Search)
	}

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	publicChatrooms := []*models.Chatroom{}
	privateChatrooms := []*models.Chatroom{}

	for _, cr := range chatrooms {
		if cr.Private {
			privateChatrooms = append(privateChatrooms, cr)
		} else {
			publicChatrooms = append(publicChatrooms, cr)
		}
	}

	data.PublicChatrooms = publicChatrooms
	data.PrivateChatrooms = privateChatrooms

	for _, room := range data.PublicChatrooms {
		names, err := app.chatroomModel.GetUsersInChatroom(room.Name, room.Private)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		formattedString := strings.Join(names, ", ")
		room.AllUsers = formattedString
	}

	for _, room := range data.PrivateChatrooms {
		names, err := app.chatroomModel.GetUsersInChatroom(room.Name, room.Private)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		formattedString := strings.Join(names, ", ")
		room.AllUsers = formattedString
	}

	app.render(w, r, http.StatusOK, "search.html", data)
}

func (app *application) chatLeavePost(w http.ResponseWriter, r *http.Request) {
	email := app.sessionManager.GetString(r.Context(), "email")
	chatroom := app.sessionManager.GetString(r.Context(), "chatroom")

	if err := app.chatroomModel.Delete(chatroom, email); err != nil {
		app.serverError(w, r, err)
		return
	}

	flash := fmt.Sprintf("Left Chatroom '%s'", chatroom)
	app.sessionManager.Put(r.Context(), "flash", flash)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) usersList(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	users, err := app.chatroomModel.GetUsersList(name)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.UsersList = users
	data.Chatroom = name
	app.render(w, r, http.StatusOK, "usersList.html", data)

}
