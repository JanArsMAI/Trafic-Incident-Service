package dto

import "time"

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
