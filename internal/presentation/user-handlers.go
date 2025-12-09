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

// AddUser godoc
// @Summary      Добавление пользователя
// @Description  Создаёт нового пользователя. Доступно только администратору.
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Param        user  body      dto.AddUserDto  true  "Данные нового пользователя"
// @Success      201   {object}  map[string]int  "ID созданного пользователя"
// @Failure      400   {object}  dto.ErrorResponse "Некорректные данные / email занят / неверная роль"
// @Failure      500   {object}  dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /users/add [post]
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

// UpdateUser godoc
// @Summary      Обновление пользователя
// @Description  Обновляет данные пользователя. Администратор может обновлять всех, обычный пользователь — только себя.
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Param user body dto.UpdateUserDto false "Новые данные пользователя"
// @Success      200   {object}  map[string]interface{} "ID обновлённого пользователя и статус"
// @Failure      400   {object}  dto.ErrorResponse "Некорректные данные / неверная роль / неверный email"
// @Failure      403   {object}  dto.ErrorResponse "Нет доступа"
// @Failure      404   {object}  dto.ErrorResponse "Пользователь не найден"
// @Failure      500   {object}  dto.ErrorResponse "Внутренняя ошибка"
// @Router       /users/update [patch]
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

// DeleteUser godoc
// @Summary      Удаление пользователя
// @Description  Удаляет пользователя по ID. Доступно только администратору.
// @Tags         Пользователи
// @Param        id   path      int  true  "ID пользователя"
// @Success      200  "Пользователь успешно удалён"
// @Failure      400  {object} dto.ErrorResponse "Некорректный ID"
// @Failure      404  {object} dto.ErrorResponse "Пользователь не найден"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка"
// @Router       /users/delete/{id} [delete]
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

// Login godoc
// @Summary      Авторизация пользователя
// @Description  Авторизует пользователя и устанавливает cookie access_token.
// @Tags         Авторизация
// @Accept       json
// @Produce      json
// @Param        credentials  body      dto.LoginDto  true  "Email и пароль"
// @Success      200          "Авторизация успешна, cookie установлена"
// @Failure      400          {object} dto.ErrorResponse "Ошибка парсинга JSON"
// @Failure      401          {object} dto.ErrorResponse "Неверный пароль"
// @Failure      404          {object} dto.ErrorResponse "Пользователь не найден"
// @Failure      500          {object} dto.ErrorResponse "Внутренняя ошибка"
// @Router       /users/login [post]
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

// GetAllUsers godoc
// @Summary Получить список всех пользователей
// @Description Возвращает пользователей постранично (через query-параметры chunk и size)
// @Tags  Пользователи
// @Accept json
// @Produce json
// @Param chunk query int true "Номер страницы (chunk)"
// @Param size query int true "Размер страницы (size)"
// @Success 200 {object} dto.UsersResponse
// @Failure 400 {object} dto.ErrorResponse "Неверные параметры"
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/get_all [get]
func (h *UserHandlers) GetAllUsers(ctx *gin.Context) {
	chunkStr := ctx.Query("chunk")
	sizeStr := ctx.Query("size")

	chunk, err := strconv.Atoi(chunkStr)
	if err != nil || chunk <= 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid 'chunk' parameter",
		})
		return
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid 'size' parameter",
		})
		return
	}

	users, err := h.svc.GetAllUsers(ctx, chunk, size)
	if err != nil {
		h.logger.Error("Get All users: error while getting", zap.Int("chunk", chunk), zap.Int("size", size), zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "failed to get users",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.UsersResponse{
		Users: users,
	})

	h.logger.Info("Get all users: successfully returned users", zap.Int("chunk", chunk), zap.Int("size", size))
}

// GetUserByName godoc
// @Summary      Получение пользователя по имени
// @Description  Возвращает данные пользователя по имени. Доступно только администратору.
// @Tags         Пользователи
// @Produce      json
// @Param        name   path      string  true  "Имя пользователя"
// @Success      200    {object}  dto.UserDto "Данные пользователя"
// @Failure      400    {object}  dto.ErrorResponse "Пустое имя"
// @Failure      404    {object}  dto.ErrorResponse "Пользователь не найден"
// @Failure      500    {object}  dto.ErrorResponse "Внутренняя ошибка"
// @Router       /users/get_user/{name} [get]
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

// Logout godoc
// @Summary      Выход из системы
// @Description  Удаляет cookie access_token и завершает пользовательскую сессию.
// @Tags         Авторизация
// @Produce      json
// @Success      200  {object}  map[string]string  "Сообщение об успешном выходе"
// @Failure      401  {object}  dto.ErrorResponse  "Необходимо выполнить вход"
// @Router       /users/logout [post]
func (h *UserHandlers) Logout(ctx *gin.Context) {
	idVal, ok := ctx.Get("user_id")
	if !ok {
		h.logger.Warn("Logout User: missing user id")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
			Message: "Forbidden to logout",
		})
		return
	}
	h.logger.Info("Logout: user logged out", zap.Any("id", idVal))
	cookie := http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	ctx.SetCookieData(&cookie)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "logged out",
	})
}
