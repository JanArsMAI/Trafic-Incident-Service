package dto

type AddUserDto struct {
	Username string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type UpdateUserDto struct {
	Username *string `json:"name"`
	Password *string `json:"password"`
	Email    *string `json:"email"`
	Role     *string `json:"role"`
}

type LoginDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
