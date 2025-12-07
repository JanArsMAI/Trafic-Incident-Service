package di

import (
	"context"
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
	repo := repos.NewPostgresUserRepo(db)
	jwtSvc := jwt.NewJwtService(cfg.Secret, time.Duration(time.Hour))
	svc := application.NewUserService(repo, jwtSvc)
	res, err := svc.AddUser(context.Background(), &dto.AddUserDto{
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
	rest.InitRoutes(r, svc, jwtSvc, logger)
	return func() {
		_ = logger.Sync()
	}
}
