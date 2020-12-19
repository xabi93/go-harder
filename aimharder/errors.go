package aimharder

import (
	"errors"
	"fmt"
)

func unknownError(err string) error {
	return fmt.Errorf("%s: %w", err, UnknownError)
}

// api errors
const (
	apiInvalidMailPassError = "Correo electrónico y/o contraseña incorrecto"
)

// Cli errors
var (
	UnknownError              = errors.New("unknown error")
	EndpointNotExistsError    = errors.New("endpoint does not exists")
	InvalidMailPassLoginError = errors.New("invalid mail or password")
	MissingAuthTokenError     = errors.New("missing auth token")
	LogoutError               = errors.New("logout")

	MissingUserIDAuthToken = errors.New("missing the user id in auth token")
	UserNotFound           = errors.New("user not found")
)
