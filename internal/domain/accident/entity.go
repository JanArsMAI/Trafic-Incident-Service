package entity

import "time"

type Accident struct {
	ID          int
	Location    string
	DateTime    time.Time
	Severity    string
	WeatherID   *int
	InspectorID *int
	CreatedAt   time.Time
}

type AccidentParticipant struct {
	ID         int
	AccidentID int
	DriverID   int
	VehicleID  int
	IsGuilty   bool
	Injuries   string
}
