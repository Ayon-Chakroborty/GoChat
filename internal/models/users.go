package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID              int
	UserName        string
	Email           string
	Hashed_Password []byte
	Created         time.Time
}

type UserModel struct{
	DB *sql.DB
}

func (m *UserModel) Insert(username, email, password string) error{
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil{
		return err
	}

	stmt := `INSERT INTO users (username, email, hashed_password, created)
	VALUES(?, ?, ?, UTC_TIMESTAMP())`

	_, err = m.DB.Exec(stmt, username, email, hashedPassword)
	if err != nil{
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError){
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email"){
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}
