package psql

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/vancho-go/lock-and-go/internal/model"
	"github.com/vancho-go/lock-and-go/pkg/customerrors"
	"regexp"
	"testing"
)

func TestDefaultUserRepository_CreateUser(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock") // Преобразование *sql.DB в *sqlx.DB

	repo := NewDefaultUserRepository(&Storage{conn: sqlxDB, log: nil})

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (username, password_hash) VALUES ($1, $2)")).
		WithArgs("testuser", "testhash").
		WillReturnResult(sqlmock.NewResult(1, 1))

	user := &model.UserHashed{
		Username:     "testuser",
		PasswordHash: "testhash",
	}

	err = repo.CreateUser(context.Background(), user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDefaultUserRepository_CreateUser_UniqueViolation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := NewDefaultUserRepository(&Storage{conn: sqlxDB, log: nil})
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (username, password_hash) VALUES ($1, $2)")).
		WithArgs("testuser", "testhash").
		WillReturnError(&pq.Error{Code: "23505", Constraint: "users_username_key"})

	user := &model.UserHashed{
		Username:     "testuser",
		PasswordHash: "testhash",
	}

	err = repo.CreateUser(context.Background(), user)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, customerrors.ErrUsernameNotUnique))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDefaultUserRepository_GetUserByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := NewDefaultUserRepository(&Storage{conn: sqlxDB, log: nil})
	rows := sqlmock.NewRows([]string{"user_id", "username", "password_hash"}).
		AddRow(1, "testuser", "testhash")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, password_hash FROM users WHERE username = $1")).
		WithArgs("testuser").
		WillReturnRows(rows)

	user, err := repo.GetUserByUsername(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "testhash", user.PasswordHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}
