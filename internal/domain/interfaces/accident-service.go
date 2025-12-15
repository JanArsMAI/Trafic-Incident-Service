package interfaces

import (
	dtoDb "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos/dto"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
)

type AccidentService interface {
	AddAccident(ctx *gin.Context, req dto.CreateAccidentDTO) (int, error)
	GetAccident(ctx *gin.Context, id int) (*dto.FullAccidentReport, error)
	UpdateAccident(ctx *gin.Context, id int, data dto.UpdateAccidentDTO) error
	UpdateWeather(ctx *gin.Context, id int, data dto.UpdateWeatherDTO) error
	AddParticipant(ctx *gin.Context, accidentID int, data dto.AddParticipantDTO) error
	UpdateParticipant(ctx *gin.Context, participantID int, data dto.UpdateParticipantDTO) error
	DeleteParticipant(ctx *gin.Context, participantID int) error
	AddParticipantViolations(ctx *gin.Context, participantID int, data dto.AddParticipantViolationsDTO) error
	DeleteParticipantViolation(ctx *gin.Context, participantID int, violationID int) error
	AddPenalty(ctx *gin.Context, data dto.AddPenaltyDTO, inspectorID int) error
	UpdatePenalty(ctx *gin.Context, penaltyID int, data dto.UpdatePenaltyDTO) error
	GetPenaltyByID(ctx *gin.Context, penaltyID int) (*dto.PenaltyResponse, error)
	GetPenaltiesByParticipant(ctx *gin.Context, participantID int) (*dto.PenaltyListResponse, error)
	AddReport(ctx *gin.Context, data dto.AddReportDTO, inspectorID int) error
	UpdateReport(ctx *gin.Context, reportID int, data dto.UpdateReportDTO) error
	GetReportByAccidentID(ctx *gin.Context, accidentID int) (*dto.ReportResponse, error)
	GetDriverAccidentStats(ctx *gin.Context) ([]dtoDb.DriverAccidentStat, error)
	GetDriverPenaltySummary(ctx *gin.Context) ([]dtoDb.DriverPenaltySummary, error)
}
