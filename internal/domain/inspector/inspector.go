package entity

import "time"

type Inspector struct {
	Id         int
	Name       string
	Number     string
	Department string
	Rank       string
	UserId     int
	CreatedAt  time.Time
}
