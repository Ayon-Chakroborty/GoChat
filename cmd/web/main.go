package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"html/template"

	"gochat.ayonchakroborty.net/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	logger    *slog.Logger
	userModel *models.UserModel
	templateCache map[string]*template.Template
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:Popcornlovers!25@/gochat?parseTime=true", "MYSQL database source name")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(*dsn)
	if err != nil{
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil{
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := application{
		logger: logger,
		userModel: &models.UserModel{DB: db},
		templateCache: templateCache,
	}

	logger.Info("starting server", "addr", *addr)

	err = http.ListenAndServe(*addr, app.routes())
	logger.Error(err.Error())
	os.Exit(1)

}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
