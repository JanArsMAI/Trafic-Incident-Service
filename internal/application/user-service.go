package application

import (
	"context"
	"errors"
	"strings"

	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/interfaces"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/user/entity"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/presentation/dto"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo       interfaces.UserRepo
	jwtService interfaces.JwtService
}

var (
	ErrUserNotFound      = errors.New("error. User is not found")
	ErrInvalidEmail      = errors.New("error. Invalid email, @ not found")
	ErrInvalidRole       = errors.New("error. Invalid role was set")
	ErrEmailIsUsed       = errors.New("error. Email is already used")
	ErrIncorrectPassword = errors.New("error. Password is incorrect")
	roleMap              = map[string]int{
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

func (u *UserService) AddUser(ctx context.Context, userDto *dto.AddUserDto) (int, error) {
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

func (u *UserService) UpdateUser(ctx context.Context, id int, dto *dto.UpdateUserDto) error {
	user, err := u.repo.GetUser(ctx, id)
	if err != nil {
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

func (u *UserService) GetUserByUsername(ctx context.Context, name string) (*dto.UserResponse, error) {
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

func (u *UserService) GetAllUsers(ctx context.Context, chunkNum, count int) ([]dto.UserResponse, error) {
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

func (u *UserService) DeleteUser(ctx context.Context, id int) error {
	err := u.repo.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func (u *UserService) Login(ctx context.Context, data dto.LoginDto) (string, error) {
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
