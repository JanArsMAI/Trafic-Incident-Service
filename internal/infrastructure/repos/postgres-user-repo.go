package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/user/entity"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos/dto"
	"github.com/jmoiron/sqlx"
)

var (
	ErrUserNotFound = errors.New("error. User with this Id or name is not found")
)

type PostgresUserRepo struct {
	db *sqlx.DB
}

func NewPostgresUserRepo(db *sqlx.DB) *PostgresUserRepo {
	return &PostgresUserRepo{
		db: db,
	}
}

func (r *PostgresUserRepo) AddUser(ctx context.Context, u *entity.User) (int, error) {
	query := `
		INSERT INTO users (username, password_hash, email, role_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id;
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		u.Username,
		u.PasswordHash,
		u.Email,
		u.RoleId,
	).Scan(&u.Id)
	if err != nil {
		return -1, err
	}
	return u.Id, nil
}

func (r *PostgresUserRepo) GetUser(ctx context.Context, id int) (*entity.User, error) {
	query := `
		SELECT id, username, password_hash, email, role_id, created_at, updated_at
		FROM users
		WHERE id = $1;
	`

	var dbUser dto.UserDto

	err := r.db.GetContext(ctx, &dbUser, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user := entity.User{
		Id:           dbUser.Id,
		Username:     dbUser.Username,
		PasswordHash: dbUser.PasswordHash,
		Email:        dbUser.Email,
		RoleId:       dbUser.Role,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}

	return &user, nil
}

func (r *PostgresUserRepo) UpdateUser(ctx context.Context, u *entity.User) error {
	query := `
		UPDATE users
		SET username = $1, password_hash = $2,email = $3,role_id = $4,updated_at = NOW()
		 WHERE id = $5;
	`
	res, err := r.db.ExecContext(ctx, query, u.Username, u.PasswordHash,
		u.Email, u.RoleId, u.Id,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *PostgresUserRepo) DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error to Delete user: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *PostgresUserRepo) GetUserByUsername(ctx context.Context, name string) (*entity.User, error) {
	query := `
		SELECT id, username, password_hash, email, role_id, created_at, updated_at
		FROM users
		WHERE username = $1;
	`

	var dbUser dto.UserDto

	err := r.db.GetContext(ctx, &dbUser, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user := entity.User{
		Id:           dbUser.Id,
		Username:     dbUser.Username,
		PasswordHash: dbUser.PasswordHash,
		Email:        dbUser.Email,
		RoleId:       dbUser.Role,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}

	return &user, nil
}

func (r *PostgresUserRepo) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, username, password_hash, email, role_id, created_at, updated_at
		FROM users
		WHERE email = $1;
	`

	var dbUser dto.UserDto

	err := r.db.GetContext(ctx, &dbUser, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user := entity.User{
		Id:           dbUser.Id,
		Username:     dbUser.Username,
		PasswordHash: dbUser.PasswordHash,
		Email:        dbUser.Email,
		RoleId:       dbUser.Role,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}

	return &user, nil
}

func (r *PostgresUserRepo) GetAll(ctx context.Context, chunk, count int) ([]entity.User, error) {
	query := `
		SELECT id, username, password_hash, email, role_id, created_at, updated_at
		FROM users
		ORDER BY id
		LIMIT $1 OFFSET $2;
	`
	var users []dto.UserDto
	offset := (chunk - 1) * count
	err := r.db.SelectContext(ctx, &users, query, count, offset)
	if err != nil {
		return nil, err
	}
	resUsers := make([]entity.User, 0, len(users))
	for _, user := range users {
		resUsers = append(resUsers, entity.User{
			Id:           user.Id,
			Username:     user.Username,
			PasswordHash: user.PasswordHash,
			Email:        user.Email,
			RoleId:       user.Role,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		})
	}
	return resUsers, nil
}

func (r *PostgresUserRepo) GetById(ctx context.Context, id int) (*entity.User, error) {
	query := `
		SELECT id, username, password_hash, email, role_id, created_at, updated_at
		FROM users
		WHERE id = $1;
	`
	var dbUser dto.UserDto
	err := r.db.GetContext(ctx, &dbUser, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	user := entity.User{
		Id:           dbUser.Id,
		Username:     dbUser.Username,
		PasswordHash: dbUser.PasswordHash,
		Email:        dbUser.Email,
		RoleId:       dbUser.Role,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}

	return &user, nil
}
