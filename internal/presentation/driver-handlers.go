package rest

import (
	"errors"
	"net/http"

	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/application"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/interfaces"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DriverHandlers struct {
	svc    interfaces.DriverService
	logger *zap.Logger
}

func NewDriverHandlers(svc interfaces.DriverService, logger *zap.Logger) *DriverHandlers {
	return &DriverHandlers{
		svc:    svc,
		logger: logger,
	}
}

// AddDriver godoc
// @Summary      Добавить нового водителя
// @Description  Создает нового водителя по переданным данным
// @Tags         Водители
// @Accept       json
// @Produce      json
// @Param        driver  body      dto.AddDriverDto  true  "Данные водителя"
// @Success      201     {object}  map[string]interface{}  "driver_id: ID созданного водителя"
// @Failure      400     {object}  dto.ErrorResponse "Некорректные данные или неверный формат даты"
// @Failure      409     {object}  dto.ErrorResponse "Водитель с таким номером лицензии уже существует"
// @Failure      500     {object}  dto.ErrorResponse "Ошибка сервера"
// @Router       /drivers/add [post]
func (h *DriverHandlers) AddDriver(ctx *gin.Context) {
	var req dto.AddDriverDto
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("AddDriver: error while parsing json", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid json body",
		})
		return
	}
	id, err := h.svc.AddDriver(ctx, req)
	if err != nil {
		h.logger.Warn("AddDriver: validation or repo error", zap.Error(err))
		switch err {
		case application.ErrIncorrectDate:
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "incorrect date format, use YYYY-MM-DD",
			})
			return

		case application.ErrBadRequest:
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "invalid or empty data",
			})
			return

		case application.ErrIncorectExperience:
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "experience must be >= 0",
			})
			return

		case application.ErrDriverWithThisLicenseIsAlreadyExists:
			ctx.AbortWithStatusJSON(http.StatusConflict, dto.ErrorResponse{
				Message: "driver with this license already exists",
			})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "failed to add driver",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":   "driver added",
		"driver_id": id,
	})
}

// UpdateDriver godoc
// @Summary      Обновить данные водителя
// @Description  Частичное обновление данных водителя по номеру лицензии
// @Tags         Водители
// @Accept       json
// @Produce      json
// @Param        driver  body      dto.UpdateDriverDto  false  "Данные для обновления"
// @Success      200
// @Failure      400     {object}  dto.ErrorResponse "Некорректные данные"
// @Failure      404     {object}  dto.ErrorResponse "Водитель не найден"
// @Failure      409     {object}  dto.ErrorResponse "Новый номер лицензии уже используется"
// @Failure      500     {object}  dto.ErrorResponse "Ошибка сервера"
// @Router       /drivers/update [patch]
func (h *DriverHandlers) UpdateDriver(ctx *gin.Context) {
	var req dto.UpdateDriverDto
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Update Driver: error while parsing json", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "invalid json body",
		})
		return
	}
	if req.License == "" {
		h.logger.Warn("Update Driver: empty license in request")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "license field is required",
		})
		return
	}
	err := h.svc.UpdateDriverInfo(ctx, req)
	if err != nil {
		switch err {
		case application.ErrDriverIsNotFound:
			h.logger.Warn("Update Driver: driver not found", zap.String("license", req.License))
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "driver not found",
			})
			return

		case application.ErrIncorrectDate:
			h.logger.Warn("Update Driver: incorrect date format")
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "incorrect date format (must be YYYY-MM-DD)",
			})
			return

		case application.ErrDriverWithThisLicenseIsAlreadyExists:
			h.logger.Warn("Update Driver: license already exists")
			ctx.AbortWithStatusJSON(http.StatusConflict, dto.ErrorResponse{
				Message: "driver with this license already exists",
			})
			return

		case application.ErrIncorectExperience:
			h.logger.Warn("Update Driver: invalid experience")
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "experience must be non-negative",
			})
			return

		default:
			h.logger.Error("Update Driver: internal server error", zap.Error(err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}
	ctx.Status(http.StatusOK)
	h.logger.Info("successfully updated driver", zap.String("driver_license", req.License))
}

// GetDriverByLicense godoc
// @Summary      Получить водителя по номеру лицензии
// @Description  Возвращает данные водителя
// @Tags         Водители
// @Accept       json
// @Produce      json
// @Param        license   path      string  true  "Номер лицензии"
// @Success      200 {object} dto.DriverResponse
// @Failure      400 {object} dto.ErrorResponse "Пустой license"
// @Failure      404 {object} dto.ErrorResponse "Водитель не найден"
// @Failure      500 {object} dto.ErrorResponse "Ошибка сервера"
// @Router       /drivers/get_by_license/{license} [get]
func (h *DriverHandlers) GetDriverByLicense(ctx *gin.Context) {
	license := ctx.Param("license")
	if license == "" {
		h.logger.Warn("Get Driver by license: empty license number")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty license",
		})
		return
	}
	driver, err := h.svc.GetDriverByLicense(ctx, license)
	if err != nil {
		if errors.Is(err, application.ErrDriverIsNotFound) {
			ctx.AbortWithStatus(http.StatusNotFound)
			h.logger.Warn("Get Driver: driver not found", zap.String("license", license))
			return
		}
		ctx.AbortWithStatus(http.StatusInternalServerError)
		h.logger.Error("Get Driver: error while getting driver by license", zap.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, dto.DriverResponse{
		Id:               driver.Id,
		Fullname:         driver.Fullname,
		DateOfBirth:      driver.DateOfBirth,
		TotalAccidents:   driver.TotalAccidents,
		License:          driver.License,
		LicenseIssueDate: driver.LicenseIssueDate,
		Experience:       driver.Experience,
		CreatedAt:        driver.CreatedAt,
	})
	h.logger.Info("Get driver by license: successfully got", zap.String("license", license))
}

// GetDriversByName godoc
// @Summary      Получить список водителей по имени
// @Description  Возвращает список всех водителей с указанным именем (частичное совпадение допускается)
// @Tags         Водители
// @Accept       json
// @Produce      json
// @Param        name   path      string  true  "Имя водителя"
// @Success      200 {object} dto.DriversResponse
// @Failure      400 {object} dto.ErrorResponse "Пустое имя"
// @Failure      500 {object} dto.ErrorResponse "Ошибка сервера"
// @Router       /drivers/get_by_name/{name} [get]
func (h *DriverHandlers) GetDriversByName(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		h.logger.Warn("Get Drivers by name: empty license number")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "empty name",
		})
		return
	}
	drivers, err := h.svc.GetDriversByName(ctx, name)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		h.logger.Error("Get Driver: error while getting driver by name", zap.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, dto.DriversResponse{
		Drivers: drivers,
	})
	h.logger.Info("successfully returned Drivers by name")
}
