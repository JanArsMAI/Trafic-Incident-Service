package rest

import (
	"net/http"

	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/interfaces"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Middleware struct {
	logger *zap.Logger
	svc    interfaces.JwtService
}

func NewMiddleware(logger *zap.Logger, svc interfaces.JwtService) *Middleware {
	return &Middleware{
		logger: logger,
		svc:    svc,
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (h *Middleware) AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		token, err := ctx.Cookie("access_token")
		if err != nil {
			h.logger.Warn("AdminMiddleware: missing access_token cookie")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Message: "token not found",
			})
			return
		}

		res, err := h.svc.ValidateToken(token)
		if err != nil {
			h.logger.Warn("AdminMiddleware: invalid token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Message: "Invalid token",
			})
			return
		}

		if res.Role != "admin" {
			h.logger.Warn("AdminMiddleware: invalid admin role")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Message: "token is invalid",
			})
			return
		}
		ctx.Set("cur_user_id", res.UserID)
		ctx.Next()
	}
}

func (h *Middleware) UserMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie("access_token")
		if err != nil {
			h.logger.Warn("UserMiddleware: missing access_token cookie")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Message: "resource not found",
			})
			return
		}
		res, err := h.svc.ValidateToken(token)
		if err != nil {
			h.logger.Warn("UserMiddleware: invalid token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Message: "Invalid token",
			})
			return
		}
		ctx.Set("user_id", res.UserID)
		ctx.Set("role", res.Role)
		ctx.Set("cur_user_id", res.UserID)
		ctx.Next()
	}
}
