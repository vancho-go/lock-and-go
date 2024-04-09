package psql

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/vancho-go/lock-and-go/internal/model"
	"regexp"
	"testing"
	"time"
)

func TestDefaultUserDataRepository_Upsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := NewDefaultUserDataRepository(&Storage{conn: sqlxDB, log: nil})

	datum := []model.UserData{
		{
			DataID:     "data1",
			UserID:     "user1",
			Data:       "some data",
			DataType:   "type1",
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
		},
	}

	query := regexp.QuoteMeta(`
        INSERT INTO user_data (data_id, user_id, data, data_type, created_at, modified_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (data_id) DO UPDATE SET
            data = EXCLUDED.data,
            modified_at = EXCLUDED.modified_at
        WHERE user_data.modified_at < EXCLUDED.modified_at;
    `)

	mock.ExpectBegin()
	mock.ExpectExec(query).WithArgs(
		datum[0].DataID, datum[0].UserID, datum[0].Data, datum[0].DataType, datum[0].CreatedAt, datum[0].ModifiedAt,
	).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.Upsert(context.Background(), datum)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDefaultUserDataRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := NewDefaultUserDataRepository(&Storage{conn: sqlxDB, log: nil})

	datum := []model.UserData{
		{
			DataID:     "data1",
			UserID:     "user1",
			ModifiedAt: time.Now(),
		},
	}

	query := regexp.QuoteMeta(`
        DELETE FROM user_data
        WHERE data_id = $1 AND user_id = $2 AND modified_at < $3;
    `)

	mock.ExpectBegin()
	mock.ExpectExec(query).WithArgs(
		datum[0].DataID, datum[0].UserID, datum[0].ModifiedAt,
	).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.Delete(context.Background(), datum)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDefaultUserDataRepository_Read(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := NewDefaultUserDataRepository(&Storage{conn: sqlxDB, log: nil})

	userID := "user1"
	rows := sqlmock.NewRows([]string{"data_id", "user_id", "data", "data_type", "created_at", "modified_at"}).
		AddRow("data1", userID, "some data", "type1", time.Now(), time.Now())

	query := regexp.QuoteMeta(`
        SELECT data_id, user_id, data, data_type, created_at, modified_at
        FROM user_data
        WHERE user_id = $1;
    `)

	mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)

	userData, err := repo.Read(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, userData, 1)
	assert.Equal(t, "data1", userData[0].DataID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
