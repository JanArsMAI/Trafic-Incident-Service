package interfaces

import (
	entityInspector "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/inspector"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/user/entity"
	"github.com/gin-gonic/gin"
)

type UserRepo interface {
	AddUser(ctx *gin.Context, u *entity.User) (int, error)
	GetUser(ctx *gin.Context, id int) (*entity.User, error)
	UpdateUser(ctx *gin.Context, u *entity.User) error
	DeleteUser(ctx *gin.Context, id int) error
	GetUserByUsername(ctx *gin.Context, name string) (*entity.User, error)
	GetUserByEmail(ctx *gin.Context, email string) (*entity.User, error)
	GetAll(ctx *gin.Context, chunk, count int) ([]entity.User, error)
	GetById(ctx *gin.Context, id int) (*entity.User, error)
	AddInspector(ctx *gin.Context, u *entityInspector.Inspector) (int, error)
	UpdateInspector(ctx *gin.Context, u *entityInspector.Inspector) error
	GetInspectorByBadge(ctx *gin.Context, number string) (*entityInspector.Inspector, error)
	GetInspectorByID(ctx *gin.Context, id int) (*entityInspector.Inspector, error)
}
