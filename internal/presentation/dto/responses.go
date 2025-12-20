package dto

import (
	"time"

	entityDriver "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/driver"
	entityInspector "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/inspector"
	entityVehicle "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/vehicle"
	entityViolation "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/violation"
	entityWeather "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/weather"
)

type UserResponse struct {
	Id        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TokenReponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type UsersResponse struct {
	Users []UserResponse `json:"users"`
}

type DriverResponse struct {
	Id               int       `json:"id"`
	Fullname         string    `json:"full_name"`
	DateOfBirth      time.Time `json:"birthdate"`
	TotalAccidents   int       `json:"accidents"`
	License          string    `json:"license_id"`
	LicenseIssueDate time.Time `json:"license_exp_date"`
	Experience       int       `json:"experience"`
	CreatedAt        time.Time `json:"created_at"`
}

type DriversResponse struct {
	Drivers []DriverResponse `json:"drivers"`
}

type VehicleResponse struct {
	Id        int       `json:"id"`
	Number    string    `json:"number"`
	Model     string    `json:"model"`
	Year      int       `json:"year"`
	Type      string    `json:"type"`
	Owner     int       `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
}

type InspectorResponse struct {
	Id         int       `json:"id"`
	Name       string    `json:"name"`
	Number     string    `json:"number"`
	Department string    `json:"department"`
	Rank       string    `json:"rank"`
	UserId     int       `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type FullAccidentReport struct {
	AccidentID int       `json:"accident_id"`
	DateTime   time.Time `json:"date_time"`
	Location   string    `json:"location"`
	Severity   string    `json:"severity"`

	WeatherID *int                   `json:"weather_id"`
	Weather   *entityWeather.Weather `json:"weather"`

	InspectorID *int                       `json:"inspector_id"`
	Inspector   *entityInspector.Inspector `json:"inspector"`

	Participants []AccidentFullParticipant `json:"participants"`
}

type AccidentFullParticipant struct {
	ParticipantID int `json:"participant_id"`
	DriverID      int `json:"driver_id"`
	VehicleID     int `json:"vehicle_id"`

	Driver  entityDriver.Driver   `json:"driver"`
	Vehicle entityVehicle.Vehicle `json:"vehicle"`

	IsGuilty bool   `json:"is_guilty"`
	Injuries string `json:"injuries"`

	Violations []entityViolation.Violation `json:"violations"`
}

type PenaltyResponse struct {
	ID            int       `json:"id"`
	ParticipantID int       `json:"participant_id"`
	Amount        float64   `json:"amount"`
	IssuedBy      *int      `json:"issued_by,omitempty"`
	IssueDate     time.Time `json:"issue_date"`
	Status        string    `json:"status"`
}

type PenaltyListResponse struct {
	Penalties []PenaltyResponse `json:"penalties"`
}

type ReportResponse struct {
	ID          int       `json:"id"`
	AccidentID  int       `json:"accident_id"`
	InspectorID *int      `json:"inspector_id,omitempty"`
	ReportText  string    `json:"report_text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DriverAccidentStatResponse struct {
	DriverID        int    `json:"driver_id"`
	FullName        string `json:"full_name"`
	ExperienceYears int    `json:"experience_years"`
	AccidentsCount  int    `json:"accidents_count"`
	GuiltyCount     int    `json:"guilty_count"`
}

type DriverPenaltySummaryResponse struct {
	DriverID       int     `json:"driver_id"`
	FullName       string  `json:"full_name"`
	PenaltiesTotal int     `json:"penalties_total"`
	TotalAmount    float64 `json:"total_amount"`
	PaidAmount     float64 `json:"paid_amount"`
	UnpaidAmount   float64 `json:"unpaid_amount"`
}
