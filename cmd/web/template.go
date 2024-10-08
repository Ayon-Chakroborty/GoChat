package main

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

type templateData struct{
	CurrentYear int
	Form any
}

func (app *application) newTemplateData(r *http.Request) templateData{
	return templateData{
		CurrentYear: time.Now().Year(),
	}
}

func newTemplateCache() (map[string]*template.Template, error){
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil{
		return nil, err
	}

	for _, page := range pages{
		name := filepath.Base(page)

		ts, err := template.New(name).ParseFiles("./ui/html/base.html")
		if err != nil{
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/partials/*.html")
		if err != nil{
			return nil, err
		}

		ts, err = ts.ParseFiles(page)

		cache[name] = ts
	}

	return cache, nil
}