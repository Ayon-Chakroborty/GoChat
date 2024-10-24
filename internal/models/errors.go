package models

import "errors"


var (
	// No matching record found in db
	ErrNoRecord = errors.New("models: no matching record found")

	// Incorrect email and/or password when trying to login
	ErrInvalidCredentials = errors.New("models: invalid credentails")
	
	// Trys to create an account with an email already in use
	ErrDuplicateEmail = errors.New("models: duplicate email")
)