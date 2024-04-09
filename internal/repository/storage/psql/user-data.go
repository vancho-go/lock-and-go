package psql

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/vancho-go/lock-and-go/internal/model"
	"github.com/vancho-go/lock-and-go/pkg/logger"
)

// UserDataUpserter методы для апсерта пользовательских данных.
type UserDataUpserter interface {
	Upsert(ctx context.Context, data []model.UserData) error
}

// UserDataReader методы для чтения пользовательских данных.
type UserDataReader interface {
	Read(ctx context.Context, userID string) ([]model.UserData, error)
}

// UserDataDeleter методы для удаления пользовательских данных.
type UserDataDeleter interface {
	Delete(ctx context.Context, data []model.UserData) error
}

// DefaultUserDataRepository тип, реализующий UserDataUpserter, UserDataReader и UserDataDeleter
type DefaultUserDataRepository struct {
	conn *sqlx.DB
	log  *logger.Logger
}

// NewDefaultUserDataRepository конструктор DefaultUserDataRepository.
func NewDefaultUserDataRepository(storage *Storage) *DefaultUserDataRepository {
	return &DefaultUserDataRepository{
		conn: storage.conn,
		log:  storage.log}
}

// Upsert добавляет/обновляет данные пользователей из БД.
func (r *DefaultUserDataRepository) Upsert(ctx context.Context, datum []model.UserData) error {
	tx, err := r.conn.BeginTxx(ctx, nil)
	if err != nil {
		r.log.Errorf("failed to begin transaction: %v", err)
		return err
	}

	for _, data := range datum {
		query := `
        INSERT INTO user_data (data_id, user_id, data, data_type, created_at, modified_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (data_id) DO UPDATE SET
            data = EXCLUDED.data,
            modified_at = EXCLUDED.modified_at
        WHERE user_data.modified_at < EXCLUDED.modified_at;
        `
		if _, err = tx.ExecContext(ctx, query, data.DataID, data.UserID, data.Data, data.DataType, data.CreatedAt, data.ModifiedAt); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				r.log.Errorf("failed to rollback after failing to upsert user data for data_id %s: %v", data.DataID, rbErr)
			}
			r.log.Errorf("failed to upsert user data for data_id %s: %v", data.DataID, err)
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		r.log.Errorf("failed to commit transaction: %v", err)
		return err
	}
	return nil
}

// Delete удаляет данные пользователей из БД.
func (r *DefaultUserDataRepository) Delete(ctx context.Context, datum []model.UserData) error {
	tx, err := r.conn.BeginTxx(ctx, nil)
	if err != nil {
		r.log.Errorf("failed to begin transaction: %v", err)
		return err
	}

	for _, data := range datum {
		query := `
        DELETE FROM user_data
        WHERE data_id = $1 AND user_id = $2 AND modified_at < $3;
        `
		if _, err = tx.ExecContext(ctx, query, data.DataID, data.UserID, data.ModifiedAt); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				r.log.Errorf("failed to rollback after failing to delete user data for data_id %s: %v", data.DataID, rbErr)
			}
			r.log.Errorf("failed to delete user data for data_id %s: %v", data.DataID, err)
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		r.log.Errorf("failed to commit transaction: %v", err)
		return err
	}

	return nil
}

// Read считывает данные пользователей из БД.
func (r *DefaultUserDataRepository) Read(ctx context.Context, userID string) ([]model.UserData, error) {
	query := `
    SELECT data_id, user_id, data, data_type, created_at, modified_at
    FROM user_data
    WHERE user_id = $1;
    `

	var userDatum []model.UserData

	err := r.conn.SelectContext(ctx, &userDatum, query, userID)
	if err != nil {
		r.log.Errorf("failed to read user data for user_id %s: %v", userID, err)
		return nil, err
	}

	return userDatum, nil
}
