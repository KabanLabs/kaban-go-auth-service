package models

import "time"

type User struct {
	ID       string
	Email    string
	PassHash []byte
	Created  time.Time
	Updated  time.Time
}
