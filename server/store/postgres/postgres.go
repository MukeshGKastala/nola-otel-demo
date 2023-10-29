package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
)

type Config struct {
	Host         string
	User         string
	Password     string
	DatabaseName string
}

func NewAndMigrate(ctx context.Context, cfg Config) (*pgx.Conn, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.DatabaseName)

	if err := runMigrations(ctx, dsn); err != nil {
		return nil, err
	}

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return conn, nil
}

func runMigrations(ctx context.Context, dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migration", "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
