package rest

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/application"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/interfaces"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandlers struct {
	svc        interfaces.UserService
	logger     *zap.Logger
	jwtService interfaces.JwtService
}

func NewUserHandlers(svc interfaces.UserService, logger *zap.Logger, jwtService interfaces.JwtService) *UserHandlers {
	return &UserHandlers{
		svc:        svc,
		logger:     logger,
		jwtService: jwtService,
	}
}

func (h *UserHandlers) AddUser(ctx *gin.Context) {
	var reqBody dto.AddUserDto
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		h.logger.Error("Add user: error parsing json", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "error while parsing json",
		})
		return
	}
	id, err := h.svc.AddUser(ctx, &reqBody)
	if err != nil {
		switch err {
		case application.ErrEmailIsUsed,
			application.ErrInvalidEmail,
			application.ErrInvalidRole:
			h.logger.Warn("Add user: Error to add user", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: err.Error(),
			})
		default:
			h.logger.Error("Add user: internal error while adding user", zap.Error(err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	h.logger.Info("Add user: User successfully created", zap.Int("id", id))
	ctx.JSON(http.StatusCreated, gin.H{
		"id": id,
	})
}

func (h *UserHandlers) Update(ctx *gin.Context) {
	var req dto.UpdateUserDto
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Update User: error parsing json", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "error while parsing json",
		})
		return
	}
	roleVal, ok := ctx.Get("role")
	if !ok {
		h.logger.Warn("Update User: missing role")
		ctx.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{
			Message: "Forbidden to update user",
		})
		return
	}
	role, ok := roleVal.(string)
	if !ok {
		h.logger.Warn("Update User: role is not string")
		ctx.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{
			Message: "Forbidden to update user",
		})
		return
	}
	idVal, ok := ctx.Get("user_id")
	if !ok {
		h.logger.Warn("Update User: missing user id")
		ctx.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{
			Message: "Forbidden to update user",
		})
		return
	}
	userID, ok := idVal.(int)
	if !ok {
		h.logger.Warn("Update User: invalid user id type")
		ctx.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{
			Message: "Forbidden to update user",
		})
		return
	}
	if role != "admin" && userID != req.Id {
		h.logger.Warn("Update User: access denied", zap.Int("req.Id", req.Id), zap.Int("userID", userID))
		ctx.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{
			Message: "Forbidden to update user",
		})
		return
	}

	err := h.svc.UpdateUser(ctx, req.Id, &req)
	if err != nil {
		switch err {
		case application.ErrEmailIsUsed,
			application.ErrInvalidEmail,
			application.ErrInvalidRole:
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})

		case application.ErrUserNotFound:
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{Message: err.Error()})

		default:
			h.logger.Error("Update user: internal error", zap.Error(err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	h.logger.Info("Update user: Successfully updated", zap.Int("id", req.Id))

	ctx.JSON(http.StatusOK, gin.H{
		"updated_id": req.Id,
		"status":     "ok",
	})
}

func (h *UserHandlers) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		h.logger.Warn("Delete user: empty id in path")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warn("Delete user: error parsing to int", zap.Error(err))
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err = h.svc.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, application.ErrUserNotFound) {
			h.logger.Warn("Delete user: no user with id", zap.Int("id", id))
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}
		h.logger.Error("Delete user: error to delete", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
	h.logger.Info("Delete user: successfully deleted user", zap.Int("id", id))
}

func (h *UserHandlers) Login(ctx *gin.Context) {
	var body dto.LoginDto
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Error("Login: error parsing json", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "error while parsing json",
		})
		return
	}
	token, err := h.svc.Login(ctx, body)
	if err != nil {
		switch err {
		case application.ErrIncorrectPassword:
			h.logger.Warn("Login: incorrect password", zap.String("email", body.Email))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Message: "incorrect password",
			})
			return
		case application.ErrUserNotFound:
			h.logger.Warn("Login: incorrect email", zap.String("email", body.Email))
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "incorrect email, user not found",
			})
			return
		default:
			h.logger.Error("Login: error while logging", zap.Error(err))
		}
	}
	cookie := http.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	ctx.SetCookieData(&cookie)
	h.logger.Info("Login: authorized user", zap.String("email", body.Email))
}

func (h *UserHandlers) GetAllUsers(ctx *gin.Context) {
	var body dto.GetAllUsersDto
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Error("Get All: error parsing json", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "error while parsing json",
		})
		return
	}
	users, err := h.svc.GetAllUsers(ctx, body.Chunk, body.Size)
	if err != nil {
		h.logger.Error("Get All users: error while getting", zap.Int("chunk", body.Chunk), zap.Int("size", body.Size))
		return
	}
	ctx.JSON(http.StatusOK, dto.UsersResponse{
		Users: users,
	})
	h.logger.Info("Get all users: successfully returned all users")
}

func (h *UserHandlers) GetUserByName(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		h.logger.Warn("Get user: empty name in path")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	user, err := h.svc.GetUserByUsername(ctx, name)
	if err != nil {
		if errors.Is(err, application.ErrUserNotFound) {
			h.logger.Warn("Get user: user is not found", zap.String("name", name))
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}
		h.logger.Error("Get user: error while getting user by name", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, user)
	h.logger.Info("Get user: successfully returned user by name")
}
