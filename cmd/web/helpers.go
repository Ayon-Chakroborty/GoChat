package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
)

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error){
	var(
		method = r.Method
		uri = r.URL.RequestURI()
		trace = string(debug.Stack())
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, r *http.Request, status int){
	http.Error(w, http.StatusText(status), status)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData){
	ts, ok := app.templateCache[page]
	if !ok{
		err := fmt.Errorf("the template %q does not exist", page)
		app.serverError(w, r, err)
		return
	}

	// test template by writing to buffer first
	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil{
		app.serverError(w, r, err)
		return
	}

	// if successfull then write to responseWriter
	w.WriteHeader(status)

	buf.WriteTo(w)
}