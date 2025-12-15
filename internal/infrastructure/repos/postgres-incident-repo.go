package repos

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	entityParticipant "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/accident"
	entity "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/inspector"
	entityPenalty "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/penalty"
	entityReport "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/report"
	entityWeather "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/weather"
	dtoDb "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos/dto"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type IncidentRepo struct {
	db *sqlx.DB
}

var (
	ErrAccidentNotFound         = errors.New("error. Accident is not found")
	ErrWeatherNotFound          = errors.New("error. weather is not found")
	ErrParticipantAlreadyExists = errors.New("participant already exists")
	ErrParticipantIsNotFound    = errors.New("participant is not found")
	ErrViolationAlreadyAssigned = errors.New("violation is already created")
	ErrPenaltyIsNotFound        = errors.New("penalty with this id is not found")
	ErrReportIsNotFound         = errors.New("reposrt with this id is not found")
)

func NewIncidentRepo(db *sqlx.DB) *IncidentRepo {
	return &IncidentRepo{
		db: db,
	}
}

func (r *IncidentRepo) withCurUserTx(ctx *gin.Context, fn func(tx *sql.Tx) error) error {
	var curUserID int
	if v := ctx.Value("cur_user_id"); v != nil {
		curUserID, _ = v.(int)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if curUserID != 0 {
		if _, err := tx.ExecContext(ctx,
			"SELECT set_config('app.current_user_id', $1::text, true)", curUserID); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *IncidentRepo) AddIncident(ctx *gin.Context, data dto.CreateAccidentDTO) (int, error) {
	var accidentID int

	err := r.withCurUserTx(ctx, func(tx *sql.Tx) error {
		var weatherID *int = nil
		if data.Weather != nil {
			w := data.Weather
			query := `
                INSERT INTO weather (temperature, precipitation, visibility, road_condition, description)
                VALUES ($1, $2, $3, $4, $5)
                RETURNING id
            `
			if err := tx.QueryRowContext(
				ctx,
				query,
				w.Temperature,
				w.Precipitation,
				w.Visibility,
				w.RoadCondition,
				w.Description,
			).Scan(&weatherID); err != nil {
				return fmt.Errorf("failed to insert weather: %w", err)
			}
		}
		a := data.Accident
		queryAccident := `
            INSERT INTO accidents (location, date_time, weather_id, inspector_id, severity)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING id
        `
		date, _ := time.Parse("2006-01-02", a.DateTime)
		if err := tx.QueryRowContext(
			ctx,
			queryAccident,
			a.Location,
			date,
			weatherID,
			a.InspectorID,
			a.Severity,
		).Scan(&accidentID); err != nil {
			return fmt.Errorf("failed to insert accident: %w", err)
		}
		for _, p := range data.Participants {
			var participantID int
			pa := p.Participant
			pa.AccidentID = accidentID
			queryPart := `
                INSERT INTO accident_participants (accident_id, driver_id, vehicle_id, is_guilty, injuries)
                VALUES ($1, $2, $3, $4, $5)
                RETURNING id
            `
			if err := tx.QueryRowContext(
				ctx,
				queryPart,
				pa.AccidentID,
				pa.DriverID,
				pa.VehicleID,
				pa.IsGuilty,
				pa.Injuries,
			).Scan(&participantID); err != nil {
				return fmt.Errorf("failed to insert participant: %w", err)
			}
			for _, violID := range p.Violations {
				queryV := `
                    INSERT INTO participant_violations (participant_id, violation_id)
                    VALUES ($1, $2)
                `
				if _, err := tx.ExecContext(ctx, queryV, participantID, violID); err != nil {
					return fmt.Errorf("failed to insert violation (violID=%d): %w", violID, err)
				}
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return accidentID, nil
}

func (r *IncidentRepo) GetAccidentById(ctx *gin.Context, id int) (*dto.FullAccidentReport, error) {
	query := `
        SELECT accident_id,
               date_time,
               location,
               severity,
               weather_id,
               weather,
               inspector_id,
               inspector,
               participants
        FROM get_full_accident_report($1)
    `

	var (
		rep              dto.FullAccidentReport
		weatherJson      []byte
		inspectorJson    []byte
		participantsJson []byte
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rep.AccidentID,
		&rep.DateTime,
		&rep.Location,
		&rep.Severity,
		&rep.WeatherID,
		&weatherJson,
		&rep.InspectorID,
		&inspectorJson,
		&participantsJson,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAccidentNotFound
		}
		return nil, err
	}

	if weatherJson != nil {
		var w entityWeather.Weather
		if err := json.Unmarshal(weatherJson, &w); err != nil {
			return nil, err
		}
		rep.Weather = &w
	}

	if inspectorJson != nil {
		var i entity.Inspector
		if err := json.Unmarshal(inspectorJson, &i); err != nil {
			return nil, err
		}
		rep.Inspector = &i
	}

	if participantsJson != nil {
		var p []dto.AccidentFullParticipant
		if err := json.Unmarshal(participantsJson, &p); err != nil {
			return nil, err
		}
		rep.Participants = p
	}

	return &rep, nil
}

func (r *IncidentRepo) UpdateAccidentData(ctx *gin.Context, id int, data dto.UpdateAccidentDTO) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {

		setParts := []string{}
		args := []any{}
		argPos := 1

		if data.Location != nil {
			setParts = append(setParts, fmt.Sprintf("location = $%d", argPos))
			args = append(args, *data.Location)
			argPos++
		}

		if data.DateTime != nil {
			t, err := time.Parse("2006-01-02", *data.DateTime)
			if err != nil {
				return fmt.Errorf("invalid date format: %w", err)
			}
			setParts = append(setParts, fmt.Sprintf("date_time = $%d", argPos))
			args = append(args, t)
			argPos++
		}

		if data.Severity != nil {
			setParts = append(setParts, fmt.Sprintf("severity = $%d", argPos))
			args = append(args, *data.Severity)
			argPos++
		}

		if data.InspectorID != nil {
			setParts = append(setParts, fmt.Sprintf("inspector_id = $%d", argPos))
			args = append(args, *data.InspectorID)
			argPos++
		}

		if len(setParts) == 0 {
			return fmt.Errorf("no fields to update")
		}

		query := fmt.Sprintf(`
            UPDATE accidents
            SET %s
            WHERE id = $%d
        `, strings.Join(setParts, ", "), argPos)

		args = append(args, id)

		result, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to update accident: %w", err)
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return ErrAccidentNotFound
		}

		return nil
	})
}

func (r *IncidentRepo) GetWeatherById(ctx *gin.Context, id int) (*entityWeather.Weather, error) {
	query := `
        SELECT id, temperature, precipitation, visibility, road_condition, description
        FROM weather
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)
	var w entityWeather.Weather
	var (
		temp   sql.NullFloat64
		precip sql.NullString
		vis    sql.NullInt32
		road   sql.NullString
		desc   sql.NullString
	)
	err := row.Scan(
		&w.ID,
		&temp,
		&precip,
		&vis,
		&road,
		&desc,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWeatherNotFound
		}
		return nil, fmt.Errorf("failed to scan weather: %w", err)
	}
	if temp.Valid {
		w.Temperature = temp.Float64
	}
	if precip.Valid {
		w.Precipitation = precip.String
	}
	if vis.Valid {
		w.Visibility = int(vis.Int32)
	}
	if road.Valid {
		w.RoadCondition = road.String
	}
	if desc.Valid {
		w.Description = desc.String
	}
	return &w, nil
}

func (r *IncidentRepo) UpdateWeather(ctx *gin.Context, id int, data entityWeather.Weather) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {

		setParts := []string{}
		args := []any{}
		argPos := 1
		if data.Temperature != 0 {
			setParts = append(setParts, fmt.Sprintf("temperature = $%d", argPos))
			args = append(args, data.Temperature)
			argPos++
		}
		if data.Precipitation != "" {
			setParts = append(setParts, fmt.Sprintf("precipitation = $%d", argPos))
			args = append(args, data.Precipitation)
			argPos++
		}
		if data.Visibility != 0 {
			setParts = append(setParts, fmt.Sprintf("visibility = $%d", argPos))
			args = append(args, data.Visibility)
			argPos++
		}
		if data.RoadCondition != "" {
			setParts = append(setParts, fmt.Sprintf("road_condition = $%d", argPos))
			args = append(args, data.RoadCondition)
			argPos++
		}
		if data.Description != "" {
			setParts = append(setParts, fmt.Sprintf("description = $%d", argPos))
			args = append(args, data.Description)
			argPos++
		}

		if len(setParts) == 0 {
			return fmt.Errorf("no fields to update")
		}

		query := fmt.Sprintf(`
            UPDATE weather
            SET %s
            WHERE id = $%d
        `, strings.Join(setParts, ", "), argPos)

		args = append(args, id)

		result, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to update weather: %w", err)
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return ErrWeatherNotFound
		}

		return nil
	})
}

func (r *IncidentRepo) AddParticipant(
	ctx *gin.Context,
	data dto.AddParticipantDTO,
	id int,
) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {

		var participantID int

		queryParticipant := `
			INSERT INTO accident_participants
				(accident_id, driver_id, vehicle_id, is_guilty, injuries)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`

		err := tx.QueryRowContext(
			ctx,
			queryParticipant,
			id,
			data.DriverID,
			data.VehicleID,
			data.IsGuilty,
			data.Injuries,
		).Scan(&participantID)

		if err != nil {
			if isUniqueViolation(err) {
				return ErrParticipantAlreadyExists
			}
			return fmt.Errorf("repo: failed to insert participant: %w", err)
		}

		for _, violationID := range data.Violations {
			queryViolation := `
				INSERT INTO participant_violations
					(participant_id, violation_id)
				VALUES ($1, $2)
			`

			if _, err := tx.ExecContext(
				ctx,
				queryViolation,
				participantID,
				violationID,
			); err != nil {
				return fmt.Errorf(
					"repo: failed to insert violation (violation_id=%d): %w",
					violationID,
					err,
				)
			}
		}

		return nil
	})
}

func (r *IncidentRepo) GetParticipantById(ctx *gin.Context, id int) (*entityParticipant.AccidentParticipant, error) {

	query := `
		SELECT
			id,
			accident_id,
			driver_id,
			vehicle_id,
			is_guilty,
			injuries
		FROM accident_participants
		WHERE id = $1
	`

	var p entityParticipant.AccidentParticipant

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.AccidentID,
		&p.DriverID,
		&p.VehicleID,
		&p.IsGuilty,
		&p.Injuries,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrParticipantIsNotFound
		}
		return nil, fmt.Errorf("repo: failed to get participant: %w", err)
	}

	return &p, nil
}

func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}

func (r *IncidentRepo) UpdateParticipant(
	ctx *gin.Context,
	participantID int,
	data dto.UpdateParticipantDTO,
) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {

		setParts := []string{}
		args := []any{}
		argPos := 1

		if data.IsGuilty != nil {
			setParts = append(setParts, fmt.Sprintf("is_guilty = $%d", argPos))
			args = append(args, *data.IsGuilty)
			argPos++
		}

		if data.Injuries != nil {
			setParts = append(setParts, fmt.Sprintf("injuries = $%d", argPos))
			args = append(args, *data.Injuries)
			argPos++
		}

		if data.DriverID != nil {
			setParts = append(setParts, fmt.Sprintf("driver_id = $%d", argPos))
			args = append(args, *data.DriverID)
			argPos++
		}

		if data.VehicleID != nil {
			setParts = append(setParts, fmt.Sprintf("vehicle_id = $%d", argPos))
			args = append(args, *data.VehicleID)
			argPos++
		}

		if len(setParts) == 0 {
			return nil
		}

		query := fmt.Sprintf(`
			UPDATE accident_participants
			SET %s
			WHERE id = $%d
		`, strings.Join(setParts, ", "), argPos)

		args = append(args, participantID)

		res, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("repo: failed to update participant: %w", err)
		}

		rows, _ := res.RowsAffected()
		if rows == 0 {
			return ErrParticipantIsNotFound
		}

		return nil
	})
}

func (r *IncidentRepo) DeleteParticipant(
	ctx *gin.Context,
	participantID int,
) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {

		query := `
			DELETE FROM accident_participants
			WHERE id = $1
		`

		res, err := tx.ExecContext(ctx, query, participantID)
		if err != nil {
			return fmt.Errorf("repo: failed to delete participant: %w", err)
		}

		rows, _ := res.RowsAffected()
		if rows == 0 {
			return ErrParticipantIsNotFound
		}

		return nil
	})
}

func (r *IncidentRepo) AddParticipantViolations(
	ctx *gin.Context,
	participantID int,
	violations []int,
) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {

		for _, violationID := range violations {
			query := `
				INSERT INTO participant_violations
					(participant_id, violation_id)
				VALUES ($1, $2)
			`

			if _, err := tx.ExecContext(
				ctx,
				query,
				participantID,
				violationID,
			); err != nil {
				if isUniqueViolation(err) {
					return ErrViolationAlreadyAssigned
				}
				return fmt.Errorf(
					"repo: failed to add violation (violation_id=%d): %w",
					violationID,
					err,
				)
			}
		}

		return nil
	})
}

func (r *IncidentRepo) DeleteParticipantViolation(
	ctx *gin.Context,
	participantID int,
	violationID int,
) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {

		query := `
			DELETE FROM participant_violations
			WHERE participant_id = $1 AND violation_id = $2
		`

		res, err := tx.ExecContext(
			ctx,
			query,
			participantID,
			violationID,
		)
		if err != nil {
			return fmt.Errorf("repo: failed to delete violation: %w", err)
		}

		rows, _ := res.RowsAffected()
		if rows == 0 {
			return ErrParticipantIsNotFound
		}

		return nil
	})
}

func (r *IncidentRepo) AddPenalty(ctx *gin.Context, data dto.AddPenaltyDTO, inspectorId int) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {

		query := `
			INSERT INTO penalties
				(participant_id, amount, issued_by)
			VALUES ($1, $2, $3)
		`

		_, err := tx.ExecContext(
			ctx,
			query,
			data.ParticipantID,
			data.Amount,
			inspectorId,
		)
		if err != nil {
			return fmt.Errorf("repo: failed to insert penalty: %w", err)
		}
		return nil
	})
}

func (r *IncidentRepo) UpdatePenalty(ctx *gin.Context, penaltyID int, data dto.UpdatePenaltyDTO) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {

		setParts := []string{}
		args := []any{}
		argPos := 1

		if data.Amount != nil {
			setParts = append(setParts, fmt.Sprintf("amount = $%d", argPos))
			args = append(args, *data.Amount)
			argPos++
		}

		if data.Status != nil {
			setParts = append(setParts, fmt.Sprintf("status = $%d", argPos))
			args = append(args, *data.Status)
			argPos++
		}

		if len(setParts) == 0 {
			return nil
		}

		query := fmt.Sprintf(`
			UPDATE penalties
			SET %s
			WHERE id = $%d
		`, strings.Join(setParts, ", "), argPos)

		args = append(args, penaltyID)

		res, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("repo: failed to update penalty: %w", err)
		}

		rows, _ := res.RowsAffected()
		if rows == 0 {
			return ErrPenaltyIsNotFound
		}

		return nil
	})
}

func (r *IncidentRepo) GetPenaltyByID(ctx *gin.Context, penaltyID int) (*entityPenalty.Penalty, error) {

	query := `
		SELECT
			id,
			participant_id,
			amount,
			issued_by,
			issue_date,
			status
		FROM penalties
		WHERE id = $1
	`

	var p entityPenalty.Penalty
	err := r.db.QueryRowContext(ctx, query, penaltyID).Scan(
		&p.ID,
		&p.ParticipantID,
		&p.Amount,
		&p.IssuedBy,
		&p.IssueDate,
		&p.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPenaltyIsNotFound
		}
		return nil, fmt.Errorf("repo: failed to get penalty: %w", err)
	}

	return &p, nil
}

func (r *IncidentRepo) GetPenaltiesByParticipant(ctx *gin.Context, participantID int) ([]entityPenalty.Penalty, error) {

	query := `
		SELECT
			id,
			participant_id,
			amount,
			issued_by,
			issue_date,
			status
		FROM penalties
		WHERE participant_id = $1
		ORDER BY issue_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, participantID)
	if err != nil {
		return nil, fmt.Errorf("repo: failed to get penalties: %w", err)
	}
	defer rows.Close()

	var penalties []entityPenalty.Penalty

	for rows.Next() {
		var p entityPenalty.Penalty
		if err := rows.Scan(
			&p.ID,
			&p.ParticipantID,
			&p.Amount,
			&p.IssuedBy,
			&p.IssueDate,
			&p.Status,
		); err != nil {
			return nil, fmt.Errorf("repo: failed to scan penalty: %w", err)
		}
		penalties = append(penalties, p)
	}

	return penalties, nil
}

func (r *IncidentRepo) AddReport(ctx *gin.Context, accidentID int, reportText string, inspectorID int) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {

		query := `
			INSERT INTO reports
				(accident_id, inspector_id, report_text)
			VALUES ($1, $2, $3)
		`

		_, err := tx.ExecContext(
			ctx,
			query,
			accidentID,
			inspectorID,
			reportText,
		)
		if err != nil {
			return fmt.Errorf("repo: failed to insert report: %w", err)
		}

		return nil
	})
}

func (r *IncidentRepo) UpdateReport(ctx *gin.Context, reportID int, reportText string) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {

		query := `
			UPDATE reports
			SET
				report_text = $1,
				updated_at = NOW()
			WHERE id = $2
		`

		res, err := tx.ExecContext(
			ctx,
			query,
			reportText,
			reportID,
		)
		if err != nil {
			return fmt.Errorf("repo: failed to update report: %w", err)
		}

		rows, _ := res.RowsAffected()
		if rows == 0 {
			return ErrReportIsNotFound
		}

		return nil
	})
}

func (r *IncidentRepo) GetReportByAccidentID(ctx *gin.Context, accidentID int) (*entityReport.Report, error) {

	query := `
		SELECT
			id,
			accident_id,
			inspector_id,
			report_text,
			created_at,
			updated_at
		FROM reports
		WHERE accident_id = $1
	`

	var report entityReport.Report
	err := r.db.QueryRowContext(ctx, query, accidentID).Scan(
		&report.ID,
		&report.AccidentID,
		&report.InspectorID,
		&report.ReportText,
		&report.CreatedAt,
		&report.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReportIsNotFound
		}
		return nil, fmt.Errorf("repo: failed to get report: %w", err)
	}

	return &report, nil
}

func (r *IncidentRepo) GetDriverAccidentStats(ctx *gin.Context) ([]dtoDb.DriverAccidentStat, error) {
	const query = `
		SELECT
			driver_id,
			full_name,
			experience_years,
			accidents_count,
			guilty_count
		FROM view_driver_accident_stats
		ORDER BY accidents_count DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]dtoDb.DriverAccidentStat, 0)

	for rows.Next() {
		var s dtoDb.DriverAccidentStat
		if err := rows.Scan(
			&s.DriverID,
			&s.FullName,
			&s.ExperienceYears,
			&s.AccidentsCount,
			&s.GuiltyCount,
		); err != nil {
			return nil, err
		}

		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *IncidentRepo) GetDriverPenaltySummary(ctx *gin.Context) ([]dtoDb.DriverPenaltySummary, error) {

	const query = `
		SELECT
			driver_id,
			full_name,
			penalties_total,
			COALESCE(total_amount, 0),
			COALESCE(paid_amount, 0),
			COALESCE(unpaid_amount, 0)
		FROM view_penalty_summary
		ORDER BY total_amount DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]dtoDb.DriverPenaltySummary, 0)

	for rows.Next() {
		var s dtoDb.DriverPenaltySummary
		if err := rows.Scan(
			&s.DriverID,
			&s.FullName,
			&s.PenaltiesTotal,
			&s.TotalAmount,
			&s.PaidAmount,
			&s.UnpaidAmount,
		); err != nil {
			return nil, err
		}

		result = append(result, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
