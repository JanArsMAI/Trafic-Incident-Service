package rest

import (
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/interfaces"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func InitRoutes(r *gin.Engine, svc interfaces.UserService, jwtSvc interfaces.JwtService, logger *zap.Logger,
	driverSvc interfaces.DriverService, accidentSvc interfaces.AccidentService) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	middleware := NewMiddleware(logger, jwtSvc)

	userHandlers := NewUserHandlers(svc, logger, jwtSvc)
	driversHandlers := NewDriverHandlers(driverSvc, logger)
	accidentHandlers := NewAccidentHandlers(logger, accidentSvc)

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

	apiInspectors := r.Group("inspectors")
	{
		apiInspectors.POST("/add", middleware.AdminMiddleware(), userHandlers.AddInspector)
		apiInspectors.PATCH("/update", middleware.AdminMiddleware(), userHandlers.UpdateInspector)
		apiInspectors.GET("/get/:id", middleware.AdminMiddleware(), userHandlers.GetInspector)
	}

	apiDrivers := r.Group("drivers")
	{
		apiDrivers.POST("/add", middleware.UserMiddleware(), driversHandlers.AddDriver)
		apiDrivers.PATCH("/update", middleware.UserMiddleware(), driversHandlers.UpdateDriver)
		apiDrivers.GET("/get_by_license/:license", middleware.UserMiddleware(), driversHandlers.GetDriverByLicense)
		apiDrivers.GET("/get_by_name/:name", middleware.UserMiddleware(), driversHandlers.GetDriversByName)
		apiDrivers.GET("/get_stats", middleware.UserMiddleware(), accidentHandlers.GetDriverAccidentStats)
		apiDrivers.GET("/get_penalties", middleware.UserMiddleware(), accidentHandlers.GetDriverPenaltySummary)
	}

	apiVehicles := r.Group("vehicles")
	{
		apiVehicles.POST("/add", middleware.UserMiddleware(), driversHandlers.AddVehicle)
		apiVehicles.GET("/vehicles/:number", middleware.UserMiddleware(), driversHandlers.GetVehicle)
		apiVehicles.PATCH("/update", middleware.UserMiddleware(), driversHandlers.UpdateVehicle)
	}

	apiAccidents := r.Group("accidents")
	{
		apiAccidents.POST("/add", middleware.UserMiddleware(), accidentHandlers.AddAccident)
		apiAccidents.GET("/get/:id", middleware.UserMiddleware(), accidentHandlers.GetAccident)
		apiAccidents.PATCH("/update/:id", middleware.UserMiddleware(), accidentHandlers.UpdateAccident)
		apiAccidents.PATCH("/update_weather/:id", middleware.UserMiddleware(), accidentHandlers.UpdateWeather)
		apiAccidents.POST("/:id/participants", middleware.UserMiddleware(), accidentHandlers.AddParticipant)
		apiAccidents.PATCH("/participants/:id", middleware.UserMiddleware(), accidentHandlers.UpdateParticipant)
		apiAccidents.DELETE("/participants/:id", middleware.UserMiddleware(), accidentHandlers.DeleteParticipant)
		apiAccidents.POST("/participants/:id/violations", middleware.UserMiddleware(), accidentHandlers.AddParticipantViolations)
		apiAccidents.DELETE("/participants/:id/violations/:violation_id", middleware.UserMiddleware(), accidentHandlers.DeleteParticipantViolation)
	}

	apiPenalties := r.Group("penalties")
	{
		apiPenalties.POST("/", middleware.UserMiddleware(), accidentHandlers.AddPenalty)
		apiPenalties.PATCH("/:id", middleware.UserMiddleware(), accidentHandlers.UpdatePenalty)
		apiPenalties.GET("/:id", middleware.UserMiddleware(), accidentHandlers.GetPenaltyByID)
		apiPenalties.GET("/:id/all", middleware.UserMiddleware(), accidentHandlers.GetPenaltiesByParticipant)
	}

	apiReports := r.Group("reports")
	{
		apiReports.POST("/", middleware.UserMiddleware(), accidentHandlers.AddReport)
		apiReports.PATCH("/:id", middleware.UserMiddleware(), accidentHandlers.UpdateReport)
		apiReports.GET("/accidents/:id", middleware.UserMiddleware(), accidentHandlers.GetReportByAccidentID)
	}
	r.Use(CORSMiddleware())
}
