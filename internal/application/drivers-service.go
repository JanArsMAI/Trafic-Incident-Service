package application

import (
	"errors"
	"fmt"
	"time"

	entity "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/driver"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
)

type DriverService struct {
	repo *repos.PostgresDriversRepo
}

func NewDriverService(repo *repos.PostgresDriversRepo) *DriverService {
	return &DriverService{
		repo: repo,
	}
}

var (
	ErrIncorrectDate                        = errors.New("error. Date is not in correct format: year-month-day")
	ErrBadRequest                           = errors.New("error. some of fields are empty or in wrong format")
	ErrIncorectExperience                   = errors.New("error. experience is negative")
	ErrDriverWithThisLicenseIsAlreadyExists = errors.New("error. Driver With this license id is already exists")
	ErrDriverIsNotFound                     = errors.New("error. Driver is not found")
)

func (s *DriverService) AddDriver(ctx *gin.Context, driver dto.AddDriverDto) (int, error) {
	if driver.Fullname == "" ||
		driver.DateOfBirth == "" ||
		driver.License == "" ||
		driver.LicenseIssueDate == "" {
		return -1, ErrBadRequest
	}
	dob, err := time.Parse("2006-01-02", driver.DateOfBirth)
	if err != nil {
		return -1, ErrIncorrectDate
	}
	if dob.After(time.Now()) {
		return -1, ErrBadRequest
	}

	license, err := time.Parse("2006-01-02", driver.LicenseIssueDate)
	if err != nil {
		return -1, ErrIncorrectDate
	}
	if license.After(time.Now()) {
		return -1, ErrBadRequest
	}

	if license.Before(dob) {
		return -1, ErrBadRequest
	}

	if driver.Experience < 0 {
		return -1, ErrIncorectExperience
	}
	_, err = s.repo.GetDriverByLicense(ctx, driver.License)
	if err == nil {
		return -1, ErrDriverWithThisLicenseIsAlreadyExists
	}
	if err != repos.ErrDriverIsNotFound {
		return -1, fmt.Errorf("error checking drivers license: %w", err)
	}

	drv := &entity.Driver{
		Fullname:         driver.Fullname,
		DateOfBirth:      dob,
		TotalAccidents:   0,
		License:          driver.License,
		LicenseIssueDate: license,
		Experience:       driver.Experience,
	}
	id, err := s.repo.AddDriver(ctx, drv)
	if err != nil {
		return -1, fmt.Errorf("failed to add driver: %w", err)
	}

	return id, nil
}

func (s *DriverService) UpdateDriverInfo(ctx *gin.Context, dto dto.UpdateDriverDto) error {
	drv, err := s.repo.GetDriverByLicense(ctx, dto.License)
	if err != nil {
		if err == repos.ErrDriverIsNotFound {
			return ErrDriverIsNotFound
		}
		return fmt.Errorf("failed to get driver by license: %w", err)
	}
	if dto.Fullname != nil {
		drv.Fullname = *dto.Fullname
	}
	if dto.DateOfBirth != nil {
		dob, err := time.Parse("2006-01-02", *dto.DateOfBirth)
		if err != nil {
			return ErrIncorrectDate
		}
		drv.DateOfBirth = dob
	}
	if dto.NewLicense != nil {
		other, err := s.repo.GetDriverByLicense(ctx, *dto.NewLicense)
		if err == nil && other.Id != drv.Id {
			return ErrDriverWithThisLicenseIsAlreadyExists
		}
		drv.License = *dto.NewLicense
	}

	if dto.LicenseIssueDate != nil {
		lid, err := time.Parse("2006-01-02", *dto.LicenseIssueDate)
		if err != nil {
			return ErrIncorrectDate
		}
		drv.LicenseIssueDate = lid
	}

	if dto.Experience != nil {
		if *dto.Experience < 0 {
			return ErrIncorectExperience
		}
		drv.Experience = *dto.Experience
	}

	return s.repo.UpdateDriver(ctx, drv)
}

func (s *DriverService) GetDriverByLicense(ctx *gin.Context, licenseNum string) (*entity.Driver, error) {
	driver, err := s.repo.GetDriverByLicense(ctx, licenseNum)
	if err != nil {
		if errors.Is(err, repos.ErrDriverIsNotFound) {
			return nil, ErrDriverIsNotFound
		}
		return nil, err
	}
	return driver, nil
}

func (s *DriverService) GetDriversByName(ctx *gin.Context, name string) ([]dto.DriverResponse, error) {
	drivers, err := s.repo.GetDriversByName(ctx, name)
	if err != nil {
		return nil, err
	}
	res := make([]dto.DriverResponse, 0, len(drivers))
	for _, driver := range drivers {
		res = append(res, dto.DriverResponse{
			Id:               driver.Id,
			Fullname:         driver.Fullname,
			DateOfBirth:      driver.DateOfBirth,
			TotalAccidents:   driver.TotalAccidents,
			License:          driver.License,
			LicenseIssueDate: driver.LicenseIssueDate,
			Experience:       driver.Experience,
			CreatedAt:        driver.CreatedAt,
		})
	}
	return res, nil
}
