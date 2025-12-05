package interfaces

import (
	"context"

	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/user/entity"
)

type UserRepo interface {
	AddUser(ctx context.Context, u *entity.User) error
	GetUser(ctx context.Context, id int) (*entity.User, error)
	UpdateUser(ctx context.Context, u *entity.User) error
	DeleteUser(ctx context.Context, id int) error
	GetUserByUsername(ctx context.Context, name string) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
}
