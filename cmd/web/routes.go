package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))

	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /about", dynamic.ThenFunc(app.about))
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))

	protected := dynamic.Append(app.requireAuthentication)

	mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogoutPost))
	mux.Handle("GET /chat", protected.ThenFunc(app.chat))
	mux.Handle("POST /chat/room", protected.ThenFunc(app.chatRoomPost))
	mux.Handle("GET /chat/room/{name}", protected.ThenFunc(app.chatRoom))
	mux.Handle("GET /chat/search", protected.ThenFunc(app.chatSearch))
	mux.Handle("POST /chat/search", protected.ThenFunc(app.chatSearchPost))
	mux.Handle("GET /user/account", protected.ThenFunc(app.userAccount))
	mux.Handle("POST /user/account", protected.ThenFunc(app.userAccountPost))
	mux.Handle("POST /user/delete", protected.ThenFunc(app.userDeletePost))

	// websocket handler
	mux.Handle("/ws", protected.ThenFunc(app.ServeWS))

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return standard.Then(mux)
}
