package interfaces

import (
	entity "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/driver"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
)

type DriverService interface {
	GetDriverByLicense(ctx *gin.Context, licenseNum string) (*entity.Driver, error)
	GetDriversByName(ctx *gin.Context, name string) ([]dto.DriverResponse, error)
	AddDriver(ctx *gin.Context, driver dto.AddDriverDto) (int, error)
	UpdateDriverInfo(ctx *gin.Context, dto dto.UpdateDriverDto) error
}
