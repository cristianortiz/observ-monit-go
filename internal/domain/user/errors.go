package user

import "errors"

// User domain errors, bussines rules errors related only to user domain

var (
	//the user repository will use this
	ErrEmailAlreadyExists = errors.New("email already exists")
	//the user repository will use this either
	ErrUserNotFound = errors.New("user not found")
	//login or maybe auth service will use this one
	ErrInvalidCredentials = errors.New("invalid or password")
)
