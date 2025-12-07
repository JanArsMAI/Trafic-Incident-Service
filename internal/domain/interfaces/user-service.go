package interfaces

import (
	"context"

	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
)

type UserService interface {
	AddUser(ctx context.Context, userDto *dto.AddUserDto) (int, error)
	UpdateUser(ctx context.Context, id int, dto *dto.UpdateUserDto) error
	GetUserByUsername(ctx context.Context, name string) (*dto.UserResponse, error)
	GetAllUsers(ctx context.Context, chunkNum, count int) ([]dto.UserResponse, error)
	DeleteUser(ctx context.Context, id int) error
	Login(ctx context.Context, data dto.LoginDto) (string, error)
}
