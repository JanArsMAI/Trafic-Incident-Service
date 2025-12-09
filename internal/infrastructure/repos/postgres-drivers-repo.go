package repos

import (
	"database/sql"
	"errors"
	"fmt"

	entityDriver "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/driver"
	entityVehicle "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/vehicle"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos/dto"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

var (
	ErrDriverIsNotFound  = errors.New("error. driver is not found")
	ErrVehicleIsNotFound = errors.New("error. Vehicle is not found")
)

type PostgresDriversRepo struct {
	db *sqlx.DB
}

func NewPostgresDriversRepo(db *sqlx.DB) *PostgresDriversRepo {
	return &PostgresDriversRepo{
		db: db,
	}
}

func (r *PostgresDriversRepo) withCurUserTx(ctx *gin.Context, fn func(tx *sql.Tx) error) error {
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

func (p *PostgresDriversRepo) AddDriver(ctx *gin.Context, driver *entityDriver.Driver) (int, error) {
	var newID int
	err := p.withCurUserTx(ctx, func(tx *sql.Tx) error {
		query := `
			INSERT INTO drivers (
				full_name,
				date_of_birth,
				total_accidents,
				license_number,
				license_issue_date,
				experience_years,
				created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
			RETURNING id;
		`
		return tx.QueryRowContext(
			ctx,
			query,
			driver.Fullname,
			driver.DateOfBirth,
			driver.TotalAccidents,
			driver.License,
			driver.LicenseIssueDate,
			driver.Experience,
		).Scan(&newID)
	})

	if err != nil {
		return -1, err
	}
	driver.Id = newID
	return newID, nil
}

func (p *PostgresDriversRepo) UpdateDriver(ctx *gin.Context, driver *entityDriver.Driver) error {
	return p.withCurUserTx(ctx, func(tx *sql.Tx) error {
		query := `
			UPDATE drivers
			SET
				full_name = $1,
				date_of_birth = $2,
				total_accidents = $3,
				license_number = $4,
				license_issue_date = $5,
				experience_years = $6
			WHERE id = $7;
		`

		res, err := tx.ExecContext(
			ctx,
			query,
			driver.Fullname,
			driver.DateOfBirth,
			driver.TotalAccidents,
			driver.License,
			driver.LicenseIssueDate,
			driver.Experience,
			driver.Id,
		)
		if err != nil {
			return fmt.Errorf("failed to update driver: %w", err)
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows: %w", err)
		}

		if rows == 0 {
			return ErrDriverIsNotFound
		}

		return nil
	})
}

func (p *PostgresDriversRepo) GetDriverByLicense(ctx *gin.Context, license string) (*entityDriver.Driver, error) {
	query := `
		SELECT 
			id,
			full_name,
			date_of_birth,
			total_accidents,
			license_number,
			license_issue_date,
			experience_years,
			created_at
		FROM drivers
		WHERE license_number = $1;
	`

	var driver dto.DriverDto
	err := p.db.GetContext(ctx, &driver, query, license)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrDriverIsNotFound
		}
		return nil, fmt.Errorf("failed to get driver by license: %w", err)
	}

	return &entityDriver.Driver{
		Id:               driver.Id,
		Fullname:         driver.Fullname,
		DateOfBirth:      driver.DateOfBirth,
		TotalAccidents:   driver.TotalAccidents,
		License:          driver.License,
		LicenseIssueDate: driver.LicenseIssueDate,
		Experience:       driver.Experience,
		CreatedAt:        driver.CreatedAt,
	}, nil
}

func (p *PostgresDriversRepo) GetDriversByName(ctx *gin.Context, name string) ([]entityDriver.Driver, error) {
	query := `
		SELECT 
			id,
			full_name,
			date_of_birth,
			total_accidents,
			license_number,
			license_issue_date,
			experience_years,
			created_at
		FROM drivers
		WHERE full_name ILIKE $1;
	`
	pattern := "%" + name + "%"
	var dtos []dto.DriverDto
	err := p.db.SelectContext(ctx, &dtos, query, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query drivers by name: %w", err)
	}
	drivers := make([]entityDriver.Driver, 0, len(dtos))
	for _, d := range dtos {
		drivers = append(drivers, entityDriver.Driver{
			Id:               d.Id,
			Fullname:         d.Fullname,
			DateOfBirth:      d.DateOfBirth,
			TotalAccidents:   d.TotalAccidents,
			License:          d.License,
			LicenseIssueDate: d.LicenseIssueDate,
			Experience:       d.Experience,
			CreatedAt:        d.CreatedAt,
		})
	}

	return drivers, nil
}

func (p *PostgresDriversRepo) AddVehicle(ctx *gin.Context, vehicle entityVehicle.Vehicle) (int, error) {
	var newID int
	err := p.withCurUserTx(ctx, func(tx *sql.Tx) error {
		query := `
		INSERT INTO vehicles(
			plate_number,
			model,
			year,
			vehicle_type,
			owner_driver_id,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id;
	`
		return tx.QueryRowContext(
			ctx,
			query,
			vehicle.Number,
			vehicle.Model,
			vehicle.Year,
			vehicle.Type,
			vehicle.Owner,
		).Scan(&newID)
	})

	if err != nil {
		return -1, err
	}
	return newID, nil
}

func (p *PostgresDriversRepo) UpdateVehicle(ctx *gin.Context, vehicle entityVehicle.Vehicle) error {
	return p.withCurUserTx(ctx, func(tx *sql.Tx) error {
		query := `
			UPDATE vehicles
			SET
				plate_number = $1,
				model = $2,
				year = $3,
				vehicle_type = $4,
				owner_driver_id = $5
			WHERE id = $6;
		`

		res, err := tx.ExecContext(
			ctx,
			query,
			vehicle.Number,
			vehicle.Model,
			vehicle.Year,
			vehicle.Type,
			vehicle.Owner,
			vehicle.Id,
		)
		if err != nil {
			return fmt.Errorf("failed to update vehicle: %w", err)
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows: %w", err)
		}

		if rows == 0 {
			return ErrVehicleIsNotFound
		}

		return nil
	})
}

func (p *PostgresDriversRepo) GetVehicleByNumber(ctx *gin.Context, number string) (*entityVehicle.Vehicle, error) {
	query := `
		SELECT
		 id,
		 plate_number,
		 model,
		 year,
		 vehicle_type,
		 owner_driver_id,
		 created_at
		 FROM vehicles 
		 WHERE plate_number=$1;
	`
	var dto dto.VehicleDto
	if err := p.db.Get(&dto, query, number); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrVehicleIsNotFound
		}
		return nil, fmt.Errorf("error to get vehicle: %w", err)
	}
	var owner int
	if dto.Owner.Valid {
		owner = int(dto.Owner.Int32)
	} else {
		owner = 0
	}
	return &entityVehicle.Vehicle{
		Id:        dto.Id,
		Number:    dto.Number,
		Model:     dto.Model,
		Year:      dto.Year,
		Type:      dto.Type,
		Owner:     owner,
		CreatedAt: dto.CreatedAt,
	}, nil
}
