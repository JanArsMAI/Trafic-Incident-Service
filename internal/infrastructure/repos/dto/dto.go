package dto

import (
	"database/sql"
	"time"

	entityAccident "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/accident"
	entityWeather "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/weather"
)

type UserDto struct {
	Id           int       `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	Email        string    `db:"email"`
	Role         int       `db:"role_id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type DriverDto struct {
	Id               int       `db:"id"`
	Fullname         string    `db:"full_name"`
	DateOfBirth      time.Time `db:"date_of_birth"`
	TotalAccidents   int       `db:"total_accidents"`
	License          string    `db:"license_number"`
	LicenseIssueDate time.Time `db:"license_issue_date"`
	Experience       int       `db:"experience_years"`
	CreatedAt        time.Time `db:"created_at"`
}

type VehicleDto struct {
	Id        int           `db:"id"`
	Number    string        `db:"plate_number"`
	Model     string        `db:"model"`
	Year      int           `db:"year"`
	Type      string        `db:"vehicle_type"`
	Owner     sql.NullInt32 `db:"owner_driver_id"`
	CreatedAt time.Time     `db:"created_at"`
}

type CreateAccidentDTO struct {
	Accident     entityAccident.Accident
	Weather      *entityWeather.Weather
	Participants []CreateParticipantDTO
}

type CreateParticipantDTO struct {
	Participant entityAccident.AccidentParticipant
	Violations  []int
}

type DriverAccidentStat struct {
	DriverID        int    `db:"driver_id"`
	FullName        string `db:"full_name"`
	ExperienceYears int    `db:"experience_years"`
	AccidentsCount  int    `db:"accidents_count"`
	GuiltyCount     int    `db:"guilty_count"`
}

type DriverPenaltySummary struct {
	DriverID       int     `db:"driver_id"`
	FullName       string  `db:"full_name"`
	PenaltiesTotal int     `db:"penalties_total"`
	TotalAmount    float64 `db:"total_amount"`
	PaidAmount     float64 `db:"paid_amount"`
	UnpaidAmount   float64 `db:"unpaid_amount"`
}
