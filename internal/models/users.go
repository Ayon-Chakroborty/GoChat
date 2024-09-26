package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID              int
	UserName        string
	Email           string
	Hashed_Password string
	Created         time.Time
}

type UserModel struct{
	DB *sql.DB
}
