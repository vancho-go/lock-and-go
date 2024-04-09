package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/vancho-go/lock-and-go/internal/model"
	"github.com/vancho-go/lock-and-go/pkg/customerrors"
	"github.com/vancho-go/lock-and-go/pkg/logger"
)

// UserRepository методы для репозитория пользователей.
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.UserHashed) error
	GetUserByUsername(ctx context.Context, username string) (*model.UserHashed, error)
}

// DefaultUserRepository тип, который реализует UserRepository.
type DefaultUserRepository struct {
	conn *sqlx.DB
	log  *logger.Logger
}

// NewDefaultUserRepository конструктор DefaultUserRepository.
func NewDefaultUserRepository(storage *Storage) *DefaultUserRepository {
	return &DefaultUserRepository{
		conn: storage.conn,
		log:  storage.log}
}

// CreateUser метод создания нового пользователя.
func (s *DefaultUserRepository) CreateUser(ctx context.Context, user *model.UserHashed) error {
	query := "INSERT INTO users (username, password_hash) VALUES ($1, $2)"

	_, err := s.conn.ExecContext(ctx, query, user.Username, user.PasswordHash)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // Код ошибки PostgresSQL для "unique_violation"
				if pqErr.Constraint == "users_username_key" {
					return customerrors.ErrUsernameNotUnique
				}
			}
		}
		return err
	}
	return nil
}

// GetUserByUsername возвращает из БД пользователя по его Username.
func (s *DefaultUserRepository) GetUserByUsername(ctx context.Context, username string) (*model.UserHashed, error) {
	var user model.UserHashed
	query := "SELECT user_id, username, password_hash FROM users WHERE username = $1"
	err := s.conn.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("error querying user by username: %w", err)
	}
	return &user, nil
}
