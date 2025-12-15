package entity

import "time"

type Penalty struct {
	ID            int
	ParticipantID int
	Amount        float64
	IssuedBy      *int
	IssueDate     time.Time
	Status        string
}
