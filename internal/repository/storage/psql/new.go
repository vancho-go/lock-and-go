package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/vancho-go/lock-and-go/pkg/logger"
)

type Storage struct {
	conn *sqlx.DB
	log  *logger.Logger
}

func New(ctx context.Context, uri string, migrationsPath string, log *logger.Logger) (*Storage, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", uri)
	if err != nil {
		return nil, fmt.Errorf("initialize: error opening database: %w", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize: error verifying database connection: %w", err)
	}

	if err = performMigrations(db.DB, migrationsPath); err != nil {
		return nil, fmt.Errorf("initialize: error performing database migrations: %w", err)
	}

	return &Storage{conn: db, log: log}, nil
}

func performMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("performMigrations: could not create database driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath), "postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("performMigrations: migration failed: %w", err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("performMigrations: an error occurred while migrating: %w", err)
	}

	return nil
}
