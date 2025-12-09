package interfaces

import (
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
)

type UserService interface {
	AddUser(ctx *gin.Context, userDto *dto.AddUserDto) (int, error)
	UpdateUser(ctx *gin.Context, id int, dto *dto.UpdateUserDto) error
	GetUserByUsername(ctx *gin.Context, name string) (*dto.UserResponse, error)
	GetAllUsers(ctx *gin.Context, chunkNum, count int) ([]dto.UserResponse, error)
	DeleteUser(ctx *gin.Context, id int) error
	Login(ctx *gin.Context, data dto.LoginDto) (string, error)
}
