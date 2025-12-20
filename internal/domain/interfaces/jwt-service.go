package interfaces

import "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/jwt"

type JwtService interface {
	GenerateToken(userId int, role string) (string, error)
	ValidateToken(tokenStr string) (*jwt.Claims, error)
}
