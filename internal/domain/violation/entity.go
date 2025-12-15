package entity

type ParticipantViolation struct {
	ID            int
	ParticipantID int
	ViolationID   int
}

type Violation struct {
	Id          int
	Description string
	Code        string
}
