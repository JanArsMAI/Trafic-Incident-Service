package rest

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/application"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/interfaces"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AccidentHandlers struct {
	logger *zap.Logger
	svc    interfaces.AccidentService
}

func NewAccidentHandlers(logger *zap.Logger, svc interfaces.AccidentService) *AccidentHandlers {
	return &AccidentHandlers{
		logger: logger,
		svc:    svc,
	}
}

// AddAccident добавляет новый дорожный инцидент.
// @Summary Добавление нового инцидента (ДТП)
// @Description Ручка позволяет инспектору зарегистрировать новый дорожный инцидент.
// Доступ разрешён только пользователям с ролью **inspector** — проверяется по значению
// @Tags ДТП
// @Accept json
// @Produce json
// @Param accident body dto.CreateAccidentDTO true "Данные нового ДТП"
// @Success 201 {object} map[string]int "ID созданного ДТП"
// @Failure 400 {object} dto.ErrorResponse "Ошибка валидации входных данных"
// @Failure 403 {string} string "Доступ запрещён. Пользователь не является инспектором"
// @Failure 404 {object} dto.ErrorResponse "Сущность, на которую есть ссылка, не найдена"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /accidents/add [post]
func (h *AccidentHandlers) AddAccident(ctx *gin.Context) {
	role, ok := ctx.Get("role")
	if !ok || role != "inspector" {
		h.logger.Warn("Add Accident: forbidden access to add incedent", zap.String("role", role.(string)))
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}
	var body dto.CreateAccidentDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Warn("Add Accidnet: error binding JSON", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "error binding json",
		})
		return
	}

	id, err := h.svc.AddAccident(ctx, body)
	if err != nil {
		h.handleAddAccidentError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"id": id,
	})
}

func (h *AccidentHandlers) handleAddAccidentError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, application.ErrEmptyLocation),
		errors.Is(err, application.ErrBadDate),
		errors.Is(err, application.ErrInvalidSeverity),
		errors.Is(err, application.ErrBadWeather),
		errors.Is(err, application.ErrNoParticipants),
		errors.Is(err, application.ErrBadParticipant),
		errors.Is(err, application.ErrDuplicateParticipant),
		errors.Is(err, application.ErrInvalidViolationID):

		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: err.Error(),
		})
		return
	case errors.Is(err, application.ErrDriverIsNotFound),
		errors.Is(err, application.ErrVehicleIsNotFound),
		errors.Is(err, application.ErrInspectorIsNotFound),
		errors.Is(err, application.ErrReferencedEntityMiss):

		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: err.Error(),
		})
		return
	default:
		h.logger.Error("AddAccident: unexpected error", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "internal server error",
		})
		return
	}
}

// GetAccident godoc
// @Summary      Получить полную информацию о ДТП
// @Description  Возвращает расширенный отчёт о ДТП с указанием погодных условий, инспектора, участников, их транспортных средств и нарушений.
// @Tags         ДТП
// @Param        id   path      int     true  "ID ДТП"
// @Produce      json
// @Success      200  {object}  dto.FullAccidentReport "Полный отчёт о ДТП"
// @Failure      400  {object}  dto.ErrorResponse       "Некорректный ID или ошибка запроса"
// @Failure      404  {object}  dto.ErrorResponse       "ДТП с указанным ID не найдено"
// @Failure      500  {object}  dto.ErrorResponse       "Внутренняя ошибка сервера"
// @Router       /accidents/get/{id} [get]
func (h *AccidentHandlers) GetAccident(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("Get Accident: empty id field", zap.String("id", id))
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Warn("Get Accident: wrong id field", zap.Error(err))
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	report, err := h.svc.GetAccident(ctx, idInt)
	if err != nil {
		if errors.Is(err, application.ErrAccidentIsNotFound) {
			h.logger.Warn("Get Accident: accident is not found", zap.String("id", id))
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}
		h.logger.Error("Get Accident: error to get accident", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	h.logger.Info("Get Accident: successfully returned accident", zap.String("id", id))
	ctx.JSON(http.StatusOK, report)
}

// UpdateAccident обновляет данные об инциденте.
// @Summary Обновление данных ДТП
// @Description Позволяет инспектору обновить сведения об уже существующем ДТП.
// @Description Обновляться могут только переданные поля (частичное обновление).
// @Tags ДТП
// @Accept json
// @Produce json
// @Param id path int true "ID инцидента"
// @Param accident body dto.UpdateAccidentDTO false "Данные для обновления ДТП"
// @Success 200 {string} string "Инцидент успешно обновлён"
// @BadRequest {object} dto.ErrorResponse "Ошибка валидации ID или тела запроса"
// @Forbidden {object} dto.ErrorResponse "Доступ запрещён (пользователь не инспектор)"
// @NotFound {object} dto.ErrorResponse "Инцидент с указанным ID не найден"
// @InternalServerError {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /accidents/update/{id} [patch]
func (h *AccidentHandlers) UpdateAccident(ctx *gin.Context) {
	role, ok := ctx.Get("role")
	if !ok || role != "inspector" {
		h.logger.Warn("Update Accident: forbidden access to add incedent", zap.String("role", role.(string)))
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}
	var body dto.UpdateAccidentDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Warn("Update Accidnet: error binding JSON", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "error binding json",
		})
		return
	}

	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("Update Accident: empty id field", zap.String("id", id))
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Warn("Update Accident: wrong id field", zap.Error(err))
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err = h.svc.UpdateAccident(ctx, idInt, body)
	if err != nil {
		if errors.Is(err, application.ErrAccidentIsNotFound) {
			h.logger.Warn("Update Accident: accident is not found", zap.String("id", id))
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}
		if errors.Is(err, application.ErrBadRequest) {
			h.logger.Warn("Update Accident: body is wrong", zap.String("id", id))
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if errors.Is(err, application.ErrInspectorIsNotFound) {
			h.logger.Warn("Update Accident: body is wrong", zap.String("id", id))
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "inspector with this id is not exists",
			})
			return
		}
		h.logger.Error("Update Accident: error to get accident", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	h.logger.Info("Update Accident: successfully updated accident", zap.String("id", id))
	ctx.Status(http.StatusOK)
}

// UpdateWeather godoc
// @Summary      Обновить данные о погоде
// @Description  Обновляет погодные условия по указанному ID. Можно передать только те поля, которые требуется изменить.
// @Tags         ДТП
// @Accept       json
// @Produce      json
// @Param        id    path      int                    true  "ID погодной записи"
// @Param        data  body      dto.UpdateWeatherDTO   true  "Данные для обновления"
// @Success      200   "Погода успешно обновлена"
// @Failure      400   {object}  dto.ErrorResponse      "Некорректные данные или отсутствуют поля для обновления"
// @Failure      404   {object}  dto.ErrorResponse      "Запись о погоде не найдена"
// @Failure      500   "Внутренняя ошибка сервера"
// @Router       /accidents/update_weather/{id} [patch]
func (h *AccidentHandlers) UpdateWeather(ctx *gin.Context) {
	var body dto.UpdateWeatherDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Warn("Update weather: error binding JSON", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid JSON",
		})
		return
	}

	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("Update weather: empty id field")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty id",
		})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Warn("Update weather: invalid id", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid id",
		})
		return
	}

	err = h.svc.UpdateWeather(ctx, idInt, body)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "no fields to update or invalid visibility",
			})
			return

		case errors.Is(err, application.ErrWeatherNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "weather record not found",
			})
			return

		default:
			h.logger.Error("Update weather: internal error", zap.Error(err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}
	h.logger.Info("successfully updated weather")
	ctx.Status(http.StatusOK)
}

// AddParticipant godoc
// @Summary Добавить участника ДТП
// @Description Добавляет нового участника в указанное ДТП
// @Tags ДТП
// @Accept json
// @Produce json
// @Param id path int true "ID ДТП"
// @Param body body dto.AddParticipantDTO true "Данные участника ДТП"
// @Success 201 "Участник успешно добавлен"
// @Failure 400 {object} dto.ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} dto.ErrorResponse "ДТП или водитель не найден"
// @Failure 409 {object} dto.ErrorResponse "Участник уже существует"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /accidents/{id}/participants [post]
func (h *AccidentHandlers) AddParticipant(ctx *gin.Context) {
	var body dto.AddParticipantDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Warn("Add participant: error binding JSON", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid JSON",
		})
		return
	}

	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("Add participant: empty id field")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty accident id",
		})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Warn("Add participant: invalid id", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid accident id",
		})
		return
	}

	if err := h.svc.AddParticipant(ctx, idInt, body); err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid participant data",
			})

		case errors.Is(err, application.ErrInvalidViolationID):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid violation id",
			})

		case errors.Is(err, application.ErrAccidentIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "accident not found",
			})

		case errors.Is(err, application.ErrDriverIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "driver not found",
			})

		case errors.Is(err, repos.ErrParticipantAlreadyExists):
			ctx.AbortWithStatusJSON(http.StatusConflict, dto.ErrorResponse{
				Message: "participant already exists in this accident",
			})

		default:
			h.logger.Error("Add participant: internal error", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "internal server error",
			})
		}
		return
	}
	h.logger.Info("participant is Added")
	ctx.Status(http.StatusCreated)
}

// UpdateParticipant godoc
// @Summary Обновить данные участника ДТП
// @Description Обновляет информацию об участнике ДТП
// @Tags ДТП
// @Accept json
// @Produce json
// @Param id path int true "ID участника ДТП"
// @Param body body dto.UpdateParticipantDTO false "Обновляемые данные участника"
// @Success 200 "Данные участника обновлены"
// @Failure 400 {object} dto.ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} dto.ErrorResponse "Участник не найден"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /accidents/participants/{id} [patch]

func (h *AccidentHandlers) UpdateParticipant(ctx *gin.Context) {
	var body dto.UpdateParticipantDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Warn("Update participant: error binding JSON", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid JSON",
		})
		return
	}

	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("Update participant: empty id field")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty participant id",
		})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Warn("Update participant: invalid id", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid participant id",
		})
		return
	}

	if err := h.svc.UpdateParticipant(ctx, idInt, body); err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid participant data",
			})

		case errors.Is(err, application.ErrParticipantIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "participant not found",
			})

		default:
			h.logger.Error("Update participant: internal error", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "internal server error",
			})
		}
		return
	}

	h.logger.Info("participant updated successfully", zap.Int("participant_id", idInt))
	ctx.Status(http.StatusOK)
}

// DeleteParticipant godoc
// @Summary Удалить участника ДТП
// @Description Удаляет участника из ДТП
// @Tags ДТП
// @Produce json
// @Param id path int true "ID участника ДТП"
// @Success 204 "Участник удалён"
// @Failure 400 {object} dto.ErrorResponse "Некорректный ID"
// @Failure 404 {object} dto.ErrorResponse "Участник не найден"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /accidents/participants/{id} [delete]
func (h *AccidentHandlers) DeleteParticipant(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("Delete participant: empty id field")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty participant id",
		})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Warn("Delete participant: invalid id", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid participant id",
		})
		return
	}

	if err := h.svc.DeleteParticipant(ctx, idInt); err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid participant id",
			})

		case errors.Is(err, application.ErrParticipantIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "participant not found",
			})

		default:
			h.logger.Error("Delete participant: internal error", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "internal server error",
			})
		}
		return
	}

	h.logger.Info("participant deleted successfully", zap.Int("participant_id", idInt))
	ctx.Status(http.StatusNoContent)
}

// AddParticipantViolations godoc
// @Summary Добавить нарушения участнику ДТП
// @Description Добавляет одно или несколько нарушений участнику ДТП
// @Tags ДТП
// @Accept json
// @Produce json
// @Param id path int true "ID участника ДТП"
// @Param body body dto.AddParticipantViolationsDTO true "Список нарушений"
// @Success 201 "Нарушения успешно добавлены"
// @Failure 400 {object} dto.ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} dto.ErrorResponse "Участник не найден"
// @Failure 409 {object} dto.ErrorResponse "Нарушение уже добавлено"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /accidents/participants/{id}/violations [post]
func (h *AccidentHandlers) AddParticipantViolations(ctx *gin.Context) {
	var body dto.AddParticipantViolationsDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Warn("Add participant violations: error binding JSON", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid JSON",
		})
		return
	}

	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("Add participant violations: empty id field")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty participant id",
		})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Warn("Add participant violations: invalid id", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid participant id",
		})
		return
	}

	if err := h.svc.AddParticipantViolations(ctx, idInt, body); err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid violations data",
			})

		case errors.Is(err, application.ErrInvalidViolationID):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid violation id",
			})

		case errors.Is(err, application.ErrViolationAlreadyExists):
			ctx.AbortWithStatusJSON(http.StatusConflict, dto.ErrorResponse{
				Message: "violation already assigned to participant",
			})

		case errors.Is(err, application.ErrParticipantIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "participant not found",
			})

		default:
			h.logger.Error("Add participant violations: internal error", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "internal server error",
			})
		}
		return
	}

	h.logger.Info("participant violations added", zap.Int("participant_id", idInt))
	ctx.Status(http.StatusCreated)
}

// DeleteParticipantViolation godoc
// @Summary Удалить нарушение у участника ДТП
// @Description Удаляет конкретное нарушение у участника ДТП
// @Tags ДТП
// @Produce json
// @Param id path int true "ID участника ДТП"
// @Param violation_id path int true "ID нарушения"
// @Success 204 "Нарушение удалено"
// @Failure 400 {object} dto.ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} dto.ErrorResponse "Участник или нарушение не найдено"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /accidents/participants/{id}/violations/{violation_id} [delete]
func (h *AccidentHandlers) DeleteParticipantViolation(ctx *gin.Context) {
	pid := ctx.Param("id")
	vid := ctx.Param("violation_id")

	if pid == "" || vid == "" {
		h.logger.Warn("Delete participant violation: empty id field")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty participant or violation id",
		})
		return
	}

	participantID, err := strconv.Atoi(pid)
	if err != nil {
		h.logger.Warn("Delete participant violation: invalid participant id", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid participant id",
		})
		return
	}

	violationID, err := strconv.Atoi(vid)
	if err != nil {
		h.logger.Warn("Delete participant violation: invalid violation id", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid violation id",
		})
		return
	}

	if err := h.svc.DeleteParticipantViolation(ctx, participantID, violationID); err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid request data",
			})

		case errors.Is(err, application.ErrParticipantIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "participant or violation not found",
			})

		default:
			h.logger.Error("Delete participant violation: internal error", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "internal server error",
			})
		}
		return
	}
	h.logger.Info(
		"participant violation deleted",
		zap.Int("participant_id", participantID),
		zap.Int("violation_id", violationID),
	)
	ctx.Status(http.StatusNoContent)
}

// AddPenalty godoc
// @Summary      Добавить штраф участнику ДТП
// @Description  Создаёт штраф для участника ДТП. Инспектор определяется по access_token.
// @Tags         Штрафы
// @Accept       json
// @Produce      json
// @Param        body  body      dto.AddPenaltyDTO  true  "Данные для создания штрафа"
// @Success      201   "Штраф успешно создан"
// @Failure      400   {object}  dto.ErrorResponse  "Некорректные данные"
// @Failure      401   {object}  dto.ErrorResponse  "Неавторизован"
// @Failure      404   {object}  dto.ErrorResponse  "Участник или инспектор не найден"
// @Failure      500   {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /penalties [post]
func (h *AccidentHandlers) AddPenalty(ctx *gin.Context) {
	var body dto.AddPenaltyDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Warn("AddPenalty: error binding JSON", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid JSON",
		})
		return
	}

	curUserID, ok := ctx.Get("user_id")
	if !ok {
		h.logger.Warn("AddPenalty: missing cur_user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}

	inspectorID, ok := curUserID.(int)
	if !ok {
		h.logger.Warn("AddPenalty: invalid cur_user_id type")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}

	if err := h.svc.AddPenalty(ctx, body, inspectorID); err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid penalty data",
			})

		case errors.Is(err, application.ErrParticipantIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "participant not found",
			})

		case errors.Is(err, application.ErrInspectorIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "inspector not found",
			})

		default:
			h.logger.Error("AddPenalty: internal error", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "internal server error",
			})
		}
		return
	}

	h.logger.Info(
		"penalty added",
		zap.Int("participant_id", body.ParticipantID),
		zap.Int("inspector_id", inspectorID),
	)

	ctx.Status(http.StatusCreated)
}

// UpdatePenalty godoc
// @Summary      Обновить штраф
// @Description  Обновляет сумму и/или статус штрафа
// @Tags         Штрафы
// @Accept       json
// @Produce      json
// @Param        id    path      int                  true  "ID штрафа"
// @Param        body  body      dto.UpdatePenaltyDTO  true  "Данные для обновления штрафа"
// @Success      200   "Штраф успешно обновлён"
// @Failure      400   {object}  dto.ErrorResponse  "Некорректные данные"
// @Failure      404   {object}  dto.ErrorResponse  "Штраф не найден"
// @Failure      500   {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /penalties/{id} [patch]
func (h *AccidentHandlers) UpdatePenalty(ctx *gin.Context) {
	var body dto.UpdatePenaltyDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Warn("UpdatePenalty: error binding JSON", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid JSON",
		})
		return
	}

	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("UpdatePenalty: empty id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty penalty id",
		})
		return
	}

	penaltyID, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Warn("UpdatePenalty: invalid id", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid penalty id",
		})
		return
	}

	if err := h.svc.UpdatePenalty(ctx, penaltyID, body); err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid penalty data",
			})

		case errors.Is(err, application.ErrPenaltyIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "penalty not found",
			})

		default:
			h.logger.Error("UpdatePenalty: internal error", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "internal server error",
			})
		}
		return
	}

	h.logger.Info("penalty updated", zap.Int("penalty_id", penaltyID))
	ctx.Status(http.StatusOK)
}

// GetPenaltyByID godoc
// @Summary      Получить штраф по ID
// @Description  Возвращает информацию о штрафе
// @Tags         Штрафы
// @Produce      json
// @Param        id   path      int  true  "ID штрафа"
// @Success      200  {object}  dto.PenaltyResponse
// @Failure      400  {object}  dto.ErrorResponse  "Некорректный ID"
// @Failure      404  {object}  dto.ErrorResponse  "Штраф не найден"
// @Failure      500  {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /penalties/{id} [get]
func (h *AccidentHandlers) GetPenaltyByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("GetPenaltyByID: empty id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty penalty id",
		})
		return
	}

	penaltyID, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Warn("GetPenaltyByID: invalid id", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid penalty id",
		})
		return
	}

	resp, err := h.svc.GetPenaltyByID(ctx, penaltyID)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid penalty id",
			})

		case errors.Is(err, application.ErrPenaltyIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "penalty not found",
			})

		default:
			h.logger.Error("GetPenaltyByID: internal error", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "internal server error",
			})
		}
		return
	}

	h.logger.Info("penalty fetched", zap.Int("penalty_id", penaltyID))
	ctx.JSON(http.StatusOK, resp)
}

// GetPenaltiesByParticipant godoc
// @Summary      Получить штрафы участника ДТП
// @Description  Возвращает список всех штрафов конкретного участника
// @Tags         Штрафы
// @Produce      json
// @Param        id   path      int  true  "ID участника ДТП"
// @Success      200  {object}  dto.PenaltyListResponse
// @Failure      400  {object}  dto.ErrorResponse  "Некорректный ID участника"
// @Failure      404  {object}  dto.ErrorResponse  "Участник не найден"
// @Failure      500  {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /penalties/{id}/all [get]
func (h *AccidentHandlers) GetPenaltiesByParticipant(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("GetPenaltiesByParticipant: empty participant id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty participant id",
		})
		return
	}

	participantID, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Warn("GetPenaltiesByParticipant: invalid id", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid participant id",
		})
		return
	}

	resp, err := h.svc.GetPenaltiesByParticipant(ctx, participantID)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid participant id",
			})

		case errors.Is(err, application.ErrParticipantIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "participant not found",
			})

		default:
			h.logger.Error("GetPenaltiesByParticipant: internal error", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "internal server error",
			})
		}
		return
	}

	h.logger.Info(
		"penalties fetched",
		zap.Int("participant_id", participantID),
		zap.Int("count", len(resp.Penalties)),
	)

	ctx.JSON(http.StatusOK, resp)
}

// AddReport godoc
// @Summary      Добавить отчёт по ДТП
// @Description  Создаёт отчёт по указанному ДТП. Отчёт привязывается к текущему инспектору.
// @Tags         Отчёты по дтп
// @Accept       json
// @Produce      json
// @Param        body  body      dto.AddReportDTO  true  "Данные для создания отчёта"
// @Success      201   "Отчёт успешно создан"
// @Failure      400   {object}  dto.ErrorResponse  "Некорректные данные запроса"
// @Failure      401   {object}  dto.ErrorResponse  "Неавторизован"
// @Failure      404   {object}  dto.ErrorResponse  "ДТП или инспектор не найдены"
// @Failure      500   {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /reports [post]
func (h *AccidentHandlers) AddReport(ctx *gin.Context) {
	var body dto.AddReportDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Warn("AddReport: error binding JSON", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid JSON",
		})
		return
	}

	curUserID, ok := ctx.Get("user_id")
	if !ok {
		h.logger.Warn("AddReport: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}

	inspectorID, ok := curUserID.(int)
	if !ok {
		h.logger.Warn("AddReport: invalid user_id type")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}

	if err := h.svc.AddReport(ctx, body, inspectorID); err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid report data",
			})

		case errors.Is(err, application.ErrAccidentIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "accident not found",
			})

		case errors.Is(err, application.ErrInspectorIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "inspector not found",
			})

		default:
			h.logger.Error("AddReport: internal error", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "internal server error",
			})
		}
		return
	}

	h.logger.Info(
		"report added",
		zap.Int("accident_id", body.AccidentID),
		zap.Int("inspector_id", inspectorID),
	)

	ctx.Status(http.StatusCreated)
}

// UpdateReport godoc
// @Summary      Обновить отчёт по ДТП
// @Description  Обновляет текст отчёта по его ID
// @Tags         Отчёты по дтп
// @Accept       json
// @Produce      json
// @Param        id    path      int                 true  "ID отчёта"
// @Param        body  body      dto.UpdateReportDTO true  "Новый текст отчёта"
// @Success      200   "Отчёт успешно обновлён"
// @Failure      400   {object}  dto.ErrorResponse  "Некорректные данные запроса"
// @Failure      404   {object}  dto.ErrorResponse  "Отчёт не найден"
// @Failure      500   {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /reports/{id} [patch]
func (h *AccidentHandlers) UpdateReport(ctx *gin.Context) {
	var body dto.UpdateReportDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Warn("UpdateReport: error binding JSON", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid JSON",
		})
		return
	}

	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("UpdateReport: empty report id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty report id",
		})
		return
	}

	reportID, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Warn("UpdateReport: invalid id", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid report id",
		})
		return
	}

	if err := h.svc.UpdateReport(ctx, reportID, body); err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid report data",
			})

		case errors.Is(err, application.ErrReportIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "report not found",
			})

		default:
			h.logger.Error("UpdateReport: internal error", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "internal server error",
			})
		}
		return
	}

	h.logger.Info("report updated", zap.Int("report_id", reportID))
	ctx.Status(http.StatusOK)
}

// GetReportByAccidentID godoc
// @Summary      Получить отчёт по ДТП
// @Description  Возвращает отчёт, связанный с указанным ДТП
// @Tags         Отчёты по дтп
// @Produce      json
// @Param        id   path      int  true  "ID ДТП"
// @Success      200  {object}  dto.ReportResponse
// @Failure      400  {object}  dto.ErrorResponse  "Некорректный ID ДТП"
// @Failure      404  {object}  dto.ErrorResponse  "Отчёт не найден"
// @Failure      500  {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /reports/accidents/{id} [get]
func (h *AccidentHandlers) GetReportByAccidentID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("GetReportByAccidentID: empty accident id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty accident id",
		})
		return
	}

	accidentID, err := strconv.Atoi(id)
	if err != nil {
		h.logger.Warn("GetReportByAccidentID: invalid id", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid accident id",
		})
		return
	}

	resp, err := h.svc.GetReportByAccidentID(ctx, accidentID)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidRequest):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid accident id",
			})

		case errors.Is(err, application.ErrReportIsNotFound):
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "report not found",
			})

		default:
			h.logger.Error("GetReportByAccidentID: internal error", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "internal server error",
			})
		}
		return
	}

	h.logger.Info("report fetched", zap.Int("accident_id", accidentID))
	ctx.JSON(http.StatusOK, resp)
}

// GetDriverAccidentStats godoc
// @Summary Получить статистику ДТП по водителям
// @Description Возвращает общую статистику по всем водителям: стаж, количество ДТП и количество ДТП, где водитель признан виновным.
// @Tags Водители
// @Produce json
// @Success 200 {array} dto.DriverAccidentStatResponse "Список статистики по водителям"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /drivers/get_stats [get]
func (h *AccidentHandlers) GetDriverAccidentStats(ctx *gin.Context) {
	stats, err := h.svc.GetDriverAccidentStats(ctx)
	if err != nil {
		h.logger.Error("GetDriverAccidentStats: internal error", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "internal server error",
		})
		return
	}

	resp := make([]dto.DriverAccidentStatResponse, 0, len(stats))
	for _, s := range stats {
		resp = append(resp, dto.DriverAccidentStatResponse{
			DriverID:        s.DriverID,
			FullName:        s.FullName,
			ExperienceYears: s.ExperienceYears,
			AccidentsCount:  s.AccidentsCount,
			GuiltyCount:     s.GuiltyCount,
		})
	}

	h.logger.Info("driver accident stats fetched", zap.Int("count", len(resp)))
	ctx.JSON(http.StatusOK, resp)
}

// GetDriverPenaltySummary godoc
// @Summary Получить сводку штрафов по водителям
// @Description Возвращает финансовую аналитику по штрафам для каждого водителя: общее количество штрафов, общую сумму, оплаченные и неоплаченные штрафы.
// @Tags Водители
// @Produce json
// @Success 200 {array} dto.DriverPenaltySummaryResponse "Сводка штрафов по водителям"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /drivers/get_penalties [get]
func (h *AccidentHandlers) GetDriverPenaltySummary(ctx *gin.Context) {
	summary, err := h.svc.GetDriverPenaltySummary(ctx)
	if err != nil {
		h.logger.Error("GetDriverPenaltySummary: internal error", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "internal server error",
		})
		return
	}

	resp := make([]dto.DriverPenaltySummaryResponse, 0, len(summary))
	for _, s := range summary {
		resp = append(resp, dto.DriverPenaltySummaryResponse{
			DriverID:       s.DriverID,
			FullName:       s.FullName,
			PenaltiesTotal: s.PenaltiesTotal,
			TotalAmount:    s.TotalAmount,
			PaidAmount:     s.PaidAmount,
			UnpaidAmount:   s.UnpaidAmount,
		})
	}

	h.logger.Info("driver penalty summary fetched", zap.Int("count", len(resp)))
	ctx.JSON(http.StatusOK, resp)
}
