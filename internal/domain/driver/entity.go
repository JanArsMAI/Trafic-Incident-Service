package entity

import "time"

type Driver struct {
	Id               int
	Fullname         string
	DateOfBirth      time.Time
	TotalAccidents   int
	License          string
	LicenseIssueDate time.Time
	Experience       int
	CreatedAt        time.Time
}
