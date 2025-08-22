package storage

import "errors"

var (
	ErrUserExists           = errors.New("User already exists")
	ErrUserNotFound         = errors.New("User not found")
	ErrAppNotFound          = errors.New("App not found")
	ErrRefreshTokenNotFound = errors.New("Refresh token not found")
)
