package rest

import (
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/interfaces"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func InitRoutes(r *gin.Engine, svc interfaces.UserService, jwtSvc interfaces.JwtService, logger *zap.Logger,
	driverSvc interfaces.DriverService) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	middleware := NewMiddleware(logger, jwtSvc)
	userHandlers := NewUserHandlers(svc, logger, jwtSvc)
	driversHandlers := NewDriverHandlers(driverSvc, logger)
	apiUsers := r.Group("users")
	{
		apiUsers.POST("/add", middleware.AdminMiddleware(), userHandlers.AddUser)
		apiUsers.PATCH("/update", middleware.UserMiddleware(), userHandlers.Update)
		apiUsers.DELETE("delete/:id", middleware.AdminMiddleware(), userHandlers.DeleteUser)
		apiUsers.POST("/login", userHandlers.Login)
		apiUsers.GET("/get_all", middleware.AdminMiddleware(), userHandlers.GetAllUsers)
		apiUsers.GET("/get_user/:name", middleware.AdminMiddleware(), userHandlers.GetUserByName)
		apiUsers.POST("/logout", middleware.UserMiddleware(), userHandlers.Logout)
	}

	apiDrivers := r.Group("drivers")
	{
		apiDrivers.POST("/add", middleware.UserMiddleware(), driversHandlers.AddDriver)
		apiDrivers.PATCH("/update", middleware.UserMiddleware(), driversHandlers.UpdateDriver)
		apiDrivers.GET("/get_by_license/:license", middleware.UserMiddleware(), driversHandlers.GetDriverByLicense)
		apiDrivers.GET("/get_by_name/:name", middleware.UserMiddleware(), driversHandlers.GetDriversByName)
	}
	r.Use(CORSMiddleware())
}
