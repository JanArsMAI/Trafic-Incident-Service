package rest

import (
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/interfaces"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func InitRoutes(r *gin.Engine, svc interfaces.UserService, jwtSvc interfaces.JwtService, logger *zap.Logger) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	h := NewUserHandlers(svc, logger, jwtSvc)
	apiUsers := r.Group("users")
	{
		apiUsers.POST("/add", h.AdminMiddleware(), h.AddUser)
		apiUsers.PATCH("/update", h.UserMiddleware(), h.Update)
		apiUsers.DELETE("delete/:id", h.AdminMiddleware(), h.DeleteUser)
		apiUsers.POST("/login", h.Login)
		apiUsers.GET("/get_all", h.AdminMiddleware(), h.GetAllUsers)
		apiUsers.GET("/get_user/:name", h.AdminMiddleware(), h.GetUserByName)
	}
	r.Use(CORSMiddleware())
}
