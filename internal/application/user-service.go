package application

import (
	"errors"
	"fmt"
	"strings"

	entityInspector "github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/inspector"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/interfaces"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/user/entity"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo       interfaces.UserRepo
	jwtService interfaces.JwtService
}

var (
	ErrUserNotFound           = errors.New("error. User is not found")
	ErrInvalidEmail           = errors.New("error. Invalid email, @ not found")
	ErrInvalidRole            = errors.New("error. Invalid role was set")
	ErrEmailIsUsed            = errors.New("error. Email is already used")
	ErrIncorrectPassword      = errors.New("error. Password is incorrect")
	ErrInspectorAlreadyExists = errors.New("error. Inspector is already exists")
	ErrIncorrectRole          = errors.New("error. User is not an inspector")
	ErrInspectorIsNotFound    = errors.New("error. Inspector is not found")
	roleMap                   = map[string]int{
		"admin":     1,
		"inspector": 2,
		"analyst":   3,
	}
	backRoleMap = map[int]string{
		1: "admin",
		2: "inspector",
		3: "analyst",
	}
)

func NewUserService(repo interfaces.UserRepo, jwtsvc interfaces.JwtService) *UserService {
	return &UserService{
		repo:       repo,
		jwtService: jwtsvc,
	}
}

func passwordToHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func comparePassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (u *UserService) AddUser(ctx *gin.Context, userDto *dto.AddUserDto) (int, error) {
	_, err := u.repo.GetUserByEmail(ctx, userDto.Email)
	if err != repos.ErrUserNotFound {
		if err == nil {
			return -1, ErrEmailIsUsed
		}
		return -1, err
	}
	if !strings.Contains(userDto.Email, "@") {
		return -1, ErrInvalidEmail
	}
	roleId, ok := roleMap[userDto.Role]
	if !ok {
		return -1, ErrInvalidRole
	}
	hashedPassword, err := passwordToHash(userDto.Password)
	if err != nil {
		return -1, err
	}
	user := &entity.User{
		Username:     userDto.Username,
		PasswordHash: hashedPassword,
		RoleId:       roleId,
		Email:        userDto.Email,
	}
	id, err := u.repo.AddUser(ctx, user)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (u *UserService) UpdateUser(ctx *gin.Context, id int, dto *dto.UpdateUserDto) error {
	user, err := u.repo.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if dto.Username != nil {
		user.Username = *dto.Username
	}
	if dto.Email != nil {
		if !strings.Contains(*dto.Email, "@") {
			return ErrInvalidEmail
		}
		user.Email = *dto.Email
	}
	if dto.Role != nil {
		roleId, ok := roleMap[*dto.Role]
		if !ok {
			return ErrInvalidRole
		}
		user.RoleId = roleId
	}
	if dto.Password != nil {
		hash, err := passwordToHash(*dto.Password)
		if err != nil {
			return err
		}
		user.PasswordHash = hash
	}

	err = u.repo.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, repos.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func (u *UserService) GetUserByUsername(ctx *gin.Context, name string) (*dto.UserResponse, error) {
	user, err := u.repo.GetUserByUsername(ctx, name)
	if err != nil {
		if errors.Is(err, repos.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	role, ok := backRoleMap[user.RoleId]
	if !ok {
		role = ""
	}
	return &dto.UserResponse{
		Id:        user.Id,
		Username:  user.Username,
		Email:     user.Email,
		Role:      role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (u *UserService) GetAllUsers(ctx *gin.Context, chunkNum, count int) ([]dto.UserResponse, error) {
	users, err := u.repo.GetAll(ctx, chunkNum, count)
	if err != nil {
		return nil, err
	}
	ans := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		role, ok := backRoleMap[user.RoleId]
		if !ok {
			role = ""
		}
		ans = append(ans, dto.UserResponse{
			Id:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			Role:      role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}
	return ans, nil
}

func (u *UserService) DeleteUser(ctx *gin.Context, id int) error {
	err := u.repo.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func (u *UserService) Login(ctx *gin.Context, data dto.LoginDto) (string, error) {
	user, err := u.repo.GetUserByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, repos.ErrUserNotFound) {
			return "", ErrUserNotFound
		}
		return "", err
	}
	if !comparePassword(user.PasswordHash, data.Password) {
		return "", ErrIncorrectPassword
	}
	role, ok := backRoleMap[user.RoleId]
	if !ok {
		role = "unknown"
	}
	token, err := u.jwtService.GenerateToken(user.Id, role)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (u *UserService) AddInspector(ctx *gin.Context, body dto.AddInspector) (int, error) {
	user, err := u.repo.GetById(ctx, body.UserId)
	if err != nil {
		if errors.Is(err, repos.ErrUserNotFound) {
			return -1, ErrUserNotFound
		}
		return -1, err
	}
	if backRoleMap[user.RoleId] != "inspector" {
		return -1, ErrIncorrectRole
	}

	if body.Name == "" || body.Number == "" || body.Department == "" || body.Rank == "" {
		return -1, ErrBadRequest
	}

	existing, err := u.repo.GetInspectorByBadge(ctx, body.Number)
	if err != nil && !errors.Is(err, repos.ErrInspectorIsNotFound) {
		return -1, err
	}
	if existing != nil {
		return -1, ErrInspectorAlreadyExists
	}

	insp := &entityInspector.Inspector{
		Name:       body.Name,
		Number:     body.Number,
		Department: body.Department,
		Rank:       body.Rank,
		UserId:     body.UserId,
	}
	id, err := u.repo.AddInspector(ctx, insp)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (u *UserService) UpdateInspector(ctx *gin.Context, body dto.UpdateInspector) error {
	if body.Id <= 0 {
		return ErrBadRequest
	}
	existing, err := u.repo.GetInspectorByID(ctx, body.Id)
	if err != nil {
		if errors.Is(err, repos.ErrInspectorIsNotFound) || errors.Is(err, ErrInspectorIsNotFound) {
			return ErrInspectorIsNotFound
		}
		return ErrBadRequest
	}
	if body.Number != nil {
		newBadge := *body.Number
		if newBadge == "" {
			return ErrBadRequest
		}
		if newBadge != existing.Number {
			other, err := u.repo.GetInspectorByBadge(ctx, newBadge)
			if err == nil && other != nil {
				if other.Id != existing.Id {
					return ErrInspectorAlreadyExists
				}
			} else if err != nil && !errors.Is(err, repos.ErrInspectorIsNotFound) && !errors.Is(err, ErrInspectorIsNotFound) {
				return fmt.Errorf("failed to check badge uniqueness: %w", err)
			}
			existing.Number = newBadge
		}
	}
	if body.Name != nil {
		if *body.Name == "" {
			return ErrBadRequest
		}
		existing.Name = *body.Name
	}
	if body.Department != nil {
		if *body.Department == "" {
			return ErrBadRequest
		}
		existing.Department = *body.Department
	}
	if body.Rank != nil {
		existing.Rank = *body.Rank
	}
	if body.UserId != nil {
		if *body.UserId <= 0 {
			return ErrBadRequest
		}
		_, err := u.repo.GetById(ctx, *body.UserId)
		if err != nil {
			if errors.Is(err, repos.ErrUserNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("failed to validate user id: %w", err)
		}
		existing.UserId = *body.UserId
	}

	if err := u.repo.UpdateInspector(ctx, existing); err != nil {
		if errors.Is(err, repos.ErrInspectorIsNotFound) {
			return ErrInspectorIsNotFound
		}
		return fmt.Errorf("failed to update inspector: %w", err)
	}
	return nil
}

func (u *UserService) GetInspector(ctx *gin.Context, id int) (*dto.InspectorResponse, error) {
	inspector, err := u.repo.GetInspectorByID(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrInspectorIsNotFound) {
			return nil, ErrInspectorIsNotFound
		}
		return nil, err
	}
	return &dto.InspectorResponse{
		Id:         inspector.Id,
		Name:       inspector.Name,
		Number:     inspector.Number,
		Department: inspector.Department,
		Rank:       inspector.Rank,
		UserId:     inspector.UserId,
		CreatedAt:  inspector.CreatedAt,
	}, nil
}
