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
