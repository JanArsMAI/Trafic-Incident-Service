package di

import (
	"time"

	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/application"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/config"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/db"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/jwt"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos"
	rest "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ConfigureApp(r *gin.Engine, logger *zap.Logger, cfg config.ServerConfig) func() {
	logger.Info("Starting configuring app...")
	db, err := db.NewPostgresConnection(db.ReadConfig())
	if err != nil {
		logger.Fatal("failed to connect to db", zap.Error(err))
	}

	userRepo := repos.NewPostgresUserRepo(db)
	driversRepo := repos.NewPostgresDriversRepo(db)

	jwtSvc := jwt.NewJwtService(cfg.Secret, time.Duration(time.Hour))
	userSvc := application.NewUserService(userRepo, jwtSvc)
	driversSvc := application.NewDriverService(driversRepo)

	res, err := userSvc.AddUser(&gin.Context{}, &dto.AddUserDto{
		Username: "Admin",
		Password: "admin",
		Role:     "admin",
		Email:    "admin@mail.com",
	})
	if err != nil {
		logger.Info("user admin is already created", zap.Error(err))
	} else {
		logger.Info("user admin is created", zap.Int("id", res))
	}
	rest.InitRoutes(r, userSvc, jwtSvc, logger, driversSvc)
	return func() {
		_ = logger.Sync()
	}
}
