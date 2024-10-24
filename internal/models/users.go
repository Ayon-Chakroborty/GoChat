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

	_, err = m.DB.Exec(stmt, username, email, string(hashedPassword))
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

func (m *UserModel) Authenticate(email, password string) (int, error){
	var id int
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password FROM users where email = ?"

	// check if account exists with this email
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil{
		if errors.Is(err, sql.ErrNoRows){
			return 0, ErrInvalidCredentials
		} else{
			return 0, err
		}
	}

	// check if the user given password matches the hashed password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil{
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword){
			return 0, ErrInvalidCredentials
		} else{
			return 0, err
		}
	}

	return id, nil
}
