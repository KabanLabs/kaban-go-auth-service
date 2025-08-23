package models

import "time"

type RefreshToken struct {
	ID        string
	Token     string
	AppID     int
	UserID    string
	ExpireAt  time.Time
	CreatedAt time.Time
	Rotated   bool
}
