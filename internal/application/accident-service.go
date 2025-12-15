package application

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/interfaces"
	entityWeather "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/weather"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos"
	dtoDb "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos/dto"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
)

type AccidentService struct {
	repo       interfaces.AccidentRepo
	driverRepo *repos.PostgresDriversRepo
	userRepo   interfaces.UserRepo
}

func NewAccidentService(repo interfaces.AccidentRepo, drRepo *repos.PostgresDriversRepo, userRepo interfaces.UserRepo) *AccidentService {
	return &AccidentService{
		repo:       repo,
		driverRepo: drRepo,
		userRepo:   userRepo,
	}
}

var (
	ErrInvalidRequest         = errors.New("error. invalid request")
	ErrInvalidSeverity        = errors.New("error. invalid severity, allowed: low, medium, high, fatal")
	ErrEmptyLocation          = errors.New("error. location is required")
	ErrBadDate                = errors.New("error. invalid date_time")
	ErrBadWeather             = errors.New("error. invalid weather data")
	ErrNoParticipants         = errors.New("error. at least one participant is required")
	ErrBadParticipant         = errors.New("error. invalid participant data")
	ErrDuplicateParticipant   = errors.New("error. duplicate participant (driver+vehicle)")
	ErrInvalidViolationID     = errors.New("error. invalid violation id")
	ErrReferencedEntityMiss   = errors.New("error. referenced entity not found")
	ErrAccidentIsNotFound     = errors.New("error. accident with this id is not found")
	ErrWeatherNotFound        = errors.New("error. weather with this id is not found")
	ErrParticipantIsNotFound  = errors.New("error participant with this id is not found")
	ErrViolationAlreadyExists = errors.New("error. Violation is already exists")
	ErrPenaltyIsNotFound      = errors.New("error. Penalty is not found")
	ErrReportIsNotFound       = errors.New("error. report not found")
)

var allowedSeverities = map[string]struct{}{
	"low":    {},
	"medium": {},
	"high":   {},
	"fatal":  {},
}

func (s *AccidentService) AddAccident(ctx *gin.Context, req dto.CreateAccidentDTO) (int, error) {
	acc := req.Accident
	acc.Location = strings.TrimSpace(acc.Location)
	if acc.Location == "" {
		return 0, ErrEmptyLocation
	}
	if acc.DateTime != "" {
		_, err := time.Parse("2006-01-02", acc.DateTime)
		if err != nil {
			return -1, ErrBadDate
		}
	}
	if _, ok := allowedSeverities[strings.ToLower(acc.Severity)]; !ok {
		return 0, ErrInvalidSeverity
	}
	if req.Weather != nil {
		w := req.Weather
		if w.Visibility < 0 {
			return 0, ErrBadRequest
		}
		w.Precipitation = strings.TrimSpace(w.Precipitation)
		w.RoadCondition = strings.TrimSpace(w.RoadCondition)
	}
	if len(req.Participants) == 0 {
		return 0, ErrNoParticipants
	}

	seen := make(map[string]struct{}, len(req.Participants))

	for _, p := range req.Participants {
		pa := p.Participant
		if pa.DriverID <= 0 || pa.VehicleID <= 0 {
			return 0, ErrBadParticipant
		}
		driver, err := s.driverRepo.GetDriverById(ctx, pa.DriverID)
		if err != nil {
			return 0, ErrDriverIsNotFound
		}
		vehicle, err := s.driverRepo.GetVehicleById(ctx, pa.VehicleID)
		if err != nil {
			return 0, ErrVehicleIsNotFound
		}
		if vehicle.Owner != 0 && vehicle.Owner != driver.Id {
			return 0, ErrBadRequest
		}
		key := fmt.Sprintf("%d:%d", pa.DriverID, pa.VehicleID)
		if _, ok := seen[key]; ok {
			return 0, ErrDuplicateParticipant
		}
		seen[key] = struct{}{}
		for _, vid := range p.Violations {
			if vid <= 0 {
				return 0, ErrInvalidViolationID
			}
		}
	}
	if acc.InspectorID != nil && *acc.InspectorID > 0 {
		if _, err := s.userRepo.GetInspectorByID(ctx, *acc.InspectorID); err != nil {
			return 0, ErrInspectorIsNotFound
		}
	}
	id, err := s.repo.AddIncident(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("repo.AddIncident failed: %w", err)
	}
	return id, nil
}

func (s *AccidentService) GetAccident(ctx *gin.Context, id int) (*dto.FullAccidentReport, error) {
	report, err := s.repo.GetAccidentById(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrAccidentNotFound) {
			return nil, ErrAccidentIsNotFound
		}
		return nil, fmt.Errorf("error to get accident: %w", err)
	}
	return report, nil
}

func (s *AccidentService) UpdateAccident(ctx *gin.Context, id int, data dto.UpdateAccidentDTO) error {

	if data.Location == nil &&
		data.DateTime == nil &&
		data.Severity == nil &&
		data.InspectorID == nil {
		return ErrBadRequest
	}
	if data.Location != nil {
		if len(*data.Location) == 0 {
			return ErrBadRequest
		}
		if len(*data.Location) > 255 {
			return ErrBadRequest
		}
	}
	if data.DateTime != nil {
		_, err := time.Parse("2006-01-02", *data.DateTime)
		if err != nil {
			return ErrBadRequest
		}
	}
	if data.Severity != nil {
		switch *data.Severity {
		case "low", "medium", "high", "fatal":
		default:
			return ErrBadRequest
		}
	}
	if data.InspectorID != nil {
		_, err := s.userRepo.GetInspectorByID(ctx, *data.InspectorID)
		if err != nil {
			if errors.Is(err, repos.ErrInspectorIsNotFound) {
				return ErrInspectorIsNotFound
			}
			return fmt.Errorf("filed to check inspector: %w", err)
		}
	}
	err := s.repo.UpdateAccidentData(ctx, id, data)
	if err != nil {
		if errors.Is(err, repos.ErrAccidentNotFound) {
			return ErrAccidentIsNotFound
		}
		return fmt.Errorf("failed to update accident: %w", err)
	}

	return nil
}

func (s *AccidentService) UpdateWeather(ctx *gin.Context, id int, data dto.UpdateWeatherDTO) error {
	if data.Temperature == nil &&
		data.Precipitation == nil &&
		data.Visibility == nil &&
		data.RoadCondition == nil &&
		data.Description == nil {
		return ErrBadRequest
	}
	if data.Visibility != nil && *data.Visibility < 0 {
		return ErrBadRequest
	}
	entity := entityWeather.Weather{
		ID: id,
	}

	if data.Temperature != nil {
		entity.Temperature = *data.Temperature
	}
	if data.Precipitation != nil {
		entity.Precipitation = *data.Precipitation
	}
	if data.Visibility != nil {
		entity.Visibility = *data.Visibility
	}
	if data.RoadCondition != nil {
		entity.RoadCondition = *data.RoadCondition
	}
	if data.Description != nil {
		entity.Description = *data.Description
	}
	err := s.repo.UpdateWeather(ctx, id, entity)
	if err != nil {
		if errors.Is(err, repos.ErrWeatherNotFound) {
			return ErrWeatherNotFound
		}
		return fmt.Errorf("service: failed to update weather: %w", err)
	}

	return nil
}

func (s *AccidentService) AddParticipant(
	ctx *gin.Context,
	accidentID int,
	data dto.AddParticipantDTO,
) error {
	if accidentID <= 0 {
		return ErrBadRequest
	}

	if data.DriverID <= 0 {
		return ErrBadRequest
	}

	if data.VehicleID <= 0 {
		return ErrBadRequest
	}
	for _, v := range data.Violations {
		if v <= 0 {
			return ErrInvalidViolationID
		}
	}

	_, err := s.repo.GetAccidentById(ctx, accidentID)
	if err != nil {
		if errors.Is(err, repos.ErrAccidentNotFound) {
			return ErrAccidentIsNotFound
		}
		return fmt.Errorf("service: failed to check accident existence: %w", err)
	}
	_, err = s.driverRepo.GetDriverById(
		ctx,
		data.DriverID,
	)
	if err != nil {
		if errors.Is(err, repos.ErrDriverIsNotFound) {
			return ErrDriverIsNotFound
		}
		return fmt.Errorf("service: failed to check participant existence: %w", err)
	}
	if err := s.repo.AddParticipant(ctx, data, accidentID); err != nil {
		return err
	}
	return nil
}

func (s *AccidentService) UpdateParticipant(
	ctx *gin.Context,
	participantID int,
	data dto.UpdateParticipantDTO,
) error {
	if participantID <= 0 {
		return ErrBadRequest
	}

	if data.DriverID != nil && *data.DriverID <= 0 {
		return ErrBadRequest
	}

	if data.VehicleID != nil && *data.VehicleID <= 0 {
		return ErrBadRequest
	}

	if data.Injuries != nil && strings.TrimSpace(*data.Injuries) == "" {
		return ErrBadRequest
	}

	if data.IsGuilty == nil &&
		data.Injuries == nil &&
		data.DriverID == nil &&
		data.VehicleID == nil {
		return ErrBadRequest
	}

	err := s.repo.UpdateParticipant(ctx, participantID, data)
	if err != nil {
		if errors.Is(err, repos.ErrParticipantIsNotFound) {
			return ErrParticipantIsNotFound
		}
		return fmt.Errorf("service: failed to update participant: %w", err)
	}

	return nil
}

func (s *AccidentService) DeleteParticipant(
	ctx *gin.Context,
	participantID int,
) error {
	if participantID <= 0 {
		return ErrBadRequest
	}
	err := s.repo.DeleteParticipant(ctx, participantID)
	if err != nil {
		if errors.Is(err, repos.ErrParticipantIsNotFound) {
			return ErrParticipantIsNotFound
		}
		return fmt.Errorf("service: failed to delete participant: %w", err)
	}

	return nil
}

func (s *AccidentService) AddParticipantViolations(
	ctx *gin.Context,
	participantID int,
	data dto.AddParticipantViolationsDTO,
) error {
	if participantID <= 0 {
		return ErrBadRequest
	}

	if len(data.Violations) == 0 {
		return ErrBadRequest
	}

	for _, v := range data.Violations {
		if v <= 0 {
			return ErrInvalidViolationID
		}
	}

	err := s.repo.AddParticipantViolations(
		ctx,
		participantID,
		data.Violations,
	)
	if err != nil {
		if errors.Is(err, repos.ErrViolationAlreadyAssigned) {
			return ErrViolationAlreadyExists
		}
		return fmt.Errorf("service: failed to add participant violations: %w", err)
	}

	return nil
}

func (s *AccidentService) DeleteParticipantViolation(ctx *gin.Context, participantID int, violationID int) error {
	if participantID <= 0 || violationID <= 0 {
		return ErrBadRequest
	}
	err := s.repo.DeleteParticipantViolation(
		ctx,
		participantID,
		violationID,
	)
	if err != nil {
		if errors.Is(err, repos.ErrParticipantIsNotFound) {
			return ErrParticipantIsNotFound
		}
		return fmt.Errorf("service: failed to delete participant violation: %w", err)
	}

	return nil
}

func (s *AccidentService) AddPenalty(ctx *gin.Context, data dto.AddPenaltyDTO, inspectorID int) error {

	if data.ParticipantID <= 0 {
		return ErrInvalidRequest
	}

	if data.Amount < 0 {
		return ErrInvalidRequest
	}

	if inspectorID <= 0 {
		return ErrInvalidRequest
	}
	_, err := s.repo.GetParticipantById(ctx, data.ParticipantID)
	if err != nil && err != repos.ErrDriverIsNotFound {
		return fmt.Errorf("failed to check existance: %w", err)
	}
	if errors.Is(err, repos.ErrInspectorIsNotFound) {
		return ErrInspectorIsNotFound
	}
	_, err = s.userRepo.GetInspectorByID(ctx, inspectorID)
	if err != nil && err != repos.ErrInspectorIsNotFound {
		return fmt.Errorf("failed to check existance: %w", err)
	}
	if errors.Is(err, repos.ErrInspectorIsNotFound) {
		return ErrInspectorIsNotFound
	}
	err = s.repo.AddPenalty(ctx, data, inspectorID)
	if err != nil {
		if errors.Is(err, ErrParticipantIsNotFound) {
			return ErrParticipantIsNotFound
		}
		return err
	}

	return nil
}

func (s *AccidentService) UpdatePenalty(ctx *gin.Context, penaltyID int, data dto.UpdatePenaltyDTO) error {

	if penaltyID <= 0 {
		return ErrInvalidRequest
	}

	if data.Amount == nil && data.Status == nil {
		return ErrInvalidRequest
	}

	if data.Amount != nil && *data.Amount < 0 {
		return ErrInvalidRequest
	}

	if data.Status != nil {
		switch *data.Status {
		case "unpaid", "paid", "canceled":
		default:
			return ErrInvalidRequest
		}
	}
	_, err := s.repo.GetPenaltyByID(ctx, penaltyID)
	if err != nil {
		if errors.Is(err, repos.ErrPenaltyIsNotFound) {
			return ErrPenaltyIsNotFound
		}
		return fmt.Errorf("failed to check penalty existence: %w", err)
	}

	err = s.repo.UpdatePenalty(ctx, penaltyID, data)
	if err != nil {
		if errors.Is(err, repos.ErrPenaltyIsNotFound) {
			return ErrPenaltyIsNotFound
		}
		return err
	}

	return nil
}

func (s *AccidentService) GetPenaltyByID(ctx *gin.Context, penaltyID int) (*dto.PenaltyResponse, error) {
	if penaltyID <= 0 {
		return nil, ErrInvalidRequest
	}

	p, err := s.repo.GetPenaltyByID(ctx, penaltyID)
	if err != nil {
		if errors.Is(err, repos.ErrPenaltyIsNotFound) {
			return nil, ErrPenaltyIsNotFound
		}
		return nil, fmt.Errorf("failed to get penalty: %w", err)
	}

	resp := &dto.PenaltyResponse{
		ID:            p.ID,
		ParticipantID: p.ParticipantID,
		Amount:        p.Amount,
		IssuedBy:      p.IssuedBy,
		IssueDate:     p.IssueDate,
		Status:        p.Status,
	}

	return resp, nil
}

func (s *AccidentService) GetPenaltiesByParticipant(ctx *gin.Context, participantID int) (*dto.PenaltyListResponse, error) {

	if participantID <= 0 {
		return nil, ErrInvalidRequest
	}
	_, err := s.repo.GetParticipantById(ctx, participantID)
	if err != nil && err != repos.ErrParticipantIsNotFound {
		return nil, fmt.Errorf("failed to check participant existence: %w", err)
	}
	if errors.Is(err, repos.ErrParticipantIsNotFound) {
		return nil, ErrParticipantIsNotFound
	}

	penalties, err := s.repo.GetPenaltiesByParticipant(ctx, participantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get penalties: %w", err)
	}

	resp := &dto.PenaltyListResponse{
		Penalties: make([]dto.PenaltyResponse, 0, len(penalties)),
	}

	for _, p := range penalties {
		resp.Penalties = append(resp.Penalties, dto.PenaltyResponse{
			ID:            p.ID,
			ParticipantID: p.ParticipantID,
			Amount:        p.Amount,
			IssuedBy:      p.IssuedBy,
			IssueDate:     p.IssueDate,
			Status:        p.Status,
		})
	}

	return resp, nil
}

func (s *AccidentService) AddReport(ctx *gin.Context, data dto.AddReportDTO, inspectorID int) error {
	if data.AccidentID <= 0 {
		return ErrInvalidRequest
	}

	if strings.TrimSpace(data.ReportText) == "" {
		return ErrInvalidRequest
	}

	if inspectorID <= 0 {
		return ErrInvalidRequest
	}
	_, err := s.repo.GetAccidentById(ctx, data.AccidentID)
	if err != nil {
		if errors.Is(err, repos.ErrAccidentNotFound) {
			return ErrAccidentIsNotFound
		}
		return fmt.Errorf("failed to check accident existence: %w", err)
	}
	_, err = s.userRepo.GetInspectorByID(ctx, inspectorID)
	if err != nil {
		if errors.Is(err, repos.ErrInspectorIsNotFound) {
			return ErrInspectorIsNotFound
		}
		return fmt.Errorf("failed to check inspector existence: %w", err)
	}
	if err := s.repo.AddReport(
		ctx,
		data.AccidentID,
		data.ReportText,
		inspectorID,
	); err != nil {
		return err
	}
	return nil
}

func (s *AccidentService) UpdateReport(ctx *gin.Context, reportID int, data dto.UpdateReportDTO) error {
	if reportID <= 0 {
		return ErrInvalidRequest
	}

	if strings.TrimSpace(data.ReportText) == "" {
		return ErrInvalidRequest
	}
	err := s.repo.UpdateReport(ctx, reportID, data.ReportText)
	if err != nil {
		if errors.Is(err, repos.ErrReportIsNotFound) {
			return ErrReportIsNotFound
		}
		return fmt.Errorf("failed to update report: %w", err)
	}

	return nil
}

func (s *AccidentService) GetReportByAccidentID(ctx *gin.Context, accidentID int) (*dto.ReportResponse, error) {
	if accidentID <= 0 {
		return nil, ErrInvalidRequest
	}

	report, err := s.repo.GetReportByAccidentID(ctx, accidentID)
	if err != nil {
		if errors.Is(err, repos.ErrReportIsNotFound) {
			return nil, ErrReportIsNotFound
		}
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	resp := &dto.ReportResponse{
		ID:          report.ID,
		AccidentID:  report.AccidentID,
		InspectorID: report.InspectorID,
		ReportText:  report.ReportText,
		CreatedAt:   report.CreatedAt,
		UpdatedAt:   report.UpdatedAt,
	}

	return resp, nil
}

func (s *AccidentService) GetDriverAccidentStats(ctx *gin.Context) ([]dtoDb.DriverAccidentStat, error) {
	stats, err := s.repo.GetDriverAccidentStats(ctx)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *AccidentService) GetDriverPenaltySummary(ctx *gin.Context) ([]dtoDb.DriverPenaltySummary, error) {
	summary, err := s.repo.GetDriverPenaltySummary(ctx)
	if err != nil {
		return nil, err
	}

	return summary, nil
}
