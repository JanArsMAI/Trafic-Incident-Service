package repos

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/domain/user/entity"
	"github.com/JanArsMAI/Trafic-Incident-Service.git/internal/infrastructure/repos/dto"
	"github.com/gin-gonic/gin"
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

func (r *PostgresUserRepo) GetUser(ctx *gin.Context, id int) (*entity.User, error) {
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

func (r *PostgresUserRepo) GetUserByUsername(ctx *gin.Context, name string) (*entity.User, error) {
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

func (r *PostgresUserRepo) GetUserByEmail(ctx *gin.Context, email string) (*entity.User, error) {
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

func (r *PostgresUserRepo) GetAll(ctx *gin.Context, chunk, count int) ([]entity.User, error) {
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

func (r *PostgresUserRepo) GetById(ctx *gin.Context, id int) (*entity.User, error) {
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

func (r *PostgresUserRepo) withCurUserTx(ctx *gin.Context, fn func(tx *sql.Tx) error) error {
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

func (r *PostgresUserRepo) UpdateUser(ctx *gin.Context, u *entity.User) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {
		query := `
            UPDATE users
            SET username = $1, password_hash = $2, email = $3, role_id = $4, updated_at = NOW()
            WHERE id = $5;
        `
		res, err := tx.ExecContext(ctx, query, u.Username, u.PasswordHash, u.Email, u.RoleId, u.Id)
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
	})
}

func (r *PostgresUserRepo) DeleteUser(ctx *gin.Context, id int) error {
	return r.withCurUserTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, id)
		if err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows: %w", err)
		}
		if rows == 0 {
			return ErrUserNotFound
		}
		return nil
	})
}

func (r *PostgresUserRepo) AddUser(ctx *gin.Context, u *entity.User) (int, error) {
	var newID int
	err := r.withCurUserTx(ctx, func(tx *sql.Tx) error {
		query := `
            INSERT INTO users (username, password_hash, email, role_id, created_at, updated_at)
            VALUES ($1, $2, $3, $4, NOW(), NOW())
            RETURNING id;
        `
		return tx.QueryRowContext(ctx, query, u.Username, u.PasswordHash, u.Email, u.RoleId).Scan(&newID)
	})
	if err != nil {
		return -1, err
	}
	u.Id = newID
	return newID, nil
}
