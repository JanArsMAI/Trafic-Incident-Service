package dto

type AddUserDto struct {
	Username string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type UpdateUserDto struct {
	Id       int     `json:"id"`
	Username *string `json:"name"`
	Password *string `json:"password"`
	Email    *string `json:"email"`
	Role     *string `json:"role"`
}

type LoginDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AddDriverDto struct {
	Fullname         string `json:"name"`
	DateOfBirth      string `json:"date_birth"`
	License          string `json:"license_number"`
	LicenseIssueDate string `json:"license_issue_date"`
	Experience       int    `json:"experience"`
}

type UpdateDriverDto struct {
	License          string  `json:"license"`
	Fullname         *string `json:"name"`
	DateOfBirth      *string `json:"date_birth"`
	NewLicense       *string `json:"new_license"`
	LicenseIssueDate *string `json:"license_issue_date"`
	Experience       *int    `json:"experience"`
}

type AddVehicleDto struct {
	Number string `json:"number"`
	Model  string `json:"model"`
	Year   int    `json:"year"`
	Type   string `json:"type"`
	Owner  int    `json:"owner"`
}

type UpdateVehicleDto struct {
	Number string  `json:"number"`
	Model  *string `json:"model"`
	Owner  *int    `json:"owner"`
	Type   *string `json:"type"`
	Year   *int    `json:"year"`
}

type AddInspector struct {
	Name       string `json:"name"`
	Number     string `json:"badge_number"`
	Department string `json:"department"`
	Rank       string `json:"rank"`
	UserId     int    `json:"userid"`
}

type UpdateInspector struct {
	Id         int     `json:"id"`
	Name       *string `json:"name"`
	Number     *string `json:"badge_number"`
	Department *string `json:"department"`
	Rank       *string `json:"rank"`
	UserId     *int    `json:"userid"`
}

type CreateAccidentDTO struct {
	Accident     AccidentDTO      `json:"accident"`
	Weather      *WeatherDTO      `json:"weather,omitempty"`
	Participants []ParticipantDTO `json:"participants"`
}

type ParticipantDTO struct {
	Participant ParticipantEntityDTO `json:"participant"`
	Violations  []int                `json:"violations"`
}

type ParticipantEntityDTO struct {
	DriverID   int    `json:"driver_id"`
	VehicleID  int    `json:"vehicle_id"`
	IsGuilty   bool   `json:"is_guilty"`
	Injuries   string `json:"injuries"`
	AccidentID int    `json:"-"`
}

type AddParticipantDTO struct {
	DriverID   int    `json:"driver_id" binding:"required"`
	VehicleID  int    `json:"vehicle_id" binding:"required"`
	IsGuilty   bool   `json:"is_guilty"`
	Injuries   string `json:"injuries"`
	Violations []int  `json:"violations"`
}

type AccidentDTO struct {
	Location    string `json:"location"`
	DateTime    string `json:"date_time"`
	Severity    string `json:"severity"`
	InspectorID *int   `json:"inspector_id"`
}

type WeatherDTO struct {
	Temperature   float64 `json:"temperature"`
	Precipitation string  `json:"precipitation"`
	Visibility    int     `json:"visibility"`
	RoadCondition string  `json:"road_condition"`
	Description   string  `json:"description"`
}

type UpdateAccidentDTO struct {
	Location    *string `json:"location,omitempty"`
	DateTime    *string `json:"date_time,omitempty"`
	Severity    *string `json:"severity,omitempty"`
	InspectorID *int    `json:"inspector_id,omitempty"`
}

type UpdateWeatherDTO struct {
	Temperature   *float64 `json:"temperature,omitempty"`
	Precipitation *string  `json:"precipitation,omitempty"`
	Visibility    *int     `json:"visibility,omitempty"`
	RoadCondition *string  `json:"road_condition,omitempty"`
	Description   *string  `json:"description,omitempty"`
}

type UpdateParticipantDTO struct {
	IsGuilty  *bool   `json:"is_guilty"`
	Injuries  *string `json:"injuries"`
	DriverID  *int    `json:"driver_id"`
	VehicleID *int    `json:"vehicle_id"`
}

type AddParticipantViolationsDTO struct {
	Violations []int `json:"violations" binding:"required"`
}

type AddPenaltyDTO struct {
	ParticipantID int     `json:"participant_id"`
	Amount        float64 `json:"amount"`
}

type UpdatePenaltyDTO struct {
	Amount *float64 `json:"amount"`
	Status *string  `json:"status"`
}
type AddReportDTO struct {
	AccidentID int    `json:"accident_id"`
	ReportText string `json:"report_text"`
}

type UpdateReportDTO struct {
	ReportText string `json:"report_text"`
}
