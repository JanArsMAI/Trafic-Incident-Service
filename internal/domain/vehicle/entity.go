package entity

import "time"

type Vehicle struct {
	Id        int
	Number    string
	Model     string
	Year      int
	Type      string
	Owner     int
	CreatedAt time.Time
}
