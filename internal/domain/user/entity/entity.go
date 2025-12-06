package entity

import "time"

type User struct {
	Id           int
	Username     string
	PasswordHash string
	Email        string
	RoleId       int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
