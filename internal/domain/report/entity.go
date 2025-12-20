package entity

import "time"

type Report struct {
	ID          int
	AccidentID  int
	InspectorID *int
	ReportText  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
