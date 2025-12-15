package interfaces

import (
	entityParticipant "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/accident"
	entityPenalty "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/penalty"
	entityReport "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/report"
	entityWeather "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/weather"
	dtoDb "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos/dto"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
)

type AccidentRepo interface {
	AddIncident(ctx *gin.Context, data dto.CreateAccidentDTO) (int, error)
	GetAccidentById(ctx *gin.Context, id int) (*dto.FullAccidentReport, error)
	UpdateAccidentData(ctx *gin.Context, id int, data dto.UpdateAccidentDTO) error
	GetWeatherById(ctx *gin.Context, id int) (*entityWeather.Weather, error)
	UpdateWeather(ctx *gin.Context, id int, data entityWeather.Weather) error
	AddParticipant(ctx *gin.Context, data dto.AddParticipantDTO, id int) error
	UpdateParticipant(ctx *gin.Context, participantID int, data dto.UpdateParticipantDTO) error
	DeleteParticipant(ctx *gin.Context, participantID int) error
	AddParticipantViolations(ctx *gin.Context, participantID int, violations []int) error
	DeleteParticipantViolation(ctx *gin.Context, participantID int, violationID int) error
	AddPenalty(ctx *gin.Context, data dto.AddPenaltyDTO, inspectorId int) error
	UpdatePenalty(ctx *gin.Context, penaltyID int, data dto.UpdatePenaltyDTO) error
	GetPenaltyByID(ctx *gin.Context, penaltyID int) (*entityPenalty.Penalty, error)
	GetPenaltiesByParticipant(ctx *gin.Context, participantID int) ([]entityPenalty.Penalty, error)
	GetParticipantById(ctx *gin.Context, id int) (*entityParticipant.AccidentParticipant, error)
	AddReport(ctx *gin.Context, accidentID int, reportText string, inspectorID int) error
	UpdateReport(ctx *gin.Context, reportID int, reportText string) error
	GetReportByAccidentID(ctx *gin.Context, accidentID int) (*entityReport.Report, error)
	GetDriverPenaltySummary(ctx *gin.Context) ([]dtoDb.DriverPenaltySummary, error)
	GetDriverAccidentStats(ctx *gin.Context) ([]dtoDb.DriverAccidentStat, error)
}
