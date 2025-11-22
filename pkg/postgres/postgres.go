package postgres

import (
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file" // источник миграций file://
	_ "github.com/jackc/pgx/v5/stdlib"                   // регистрирует драйвер pgx для database/sql

	"github.com/jmoiron/sqlx"
)


type PDB struct {
	DB *sqlx.DB
}

func NewPDB(dsn string) (*PDB, error) {

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to connect: %w", err)
	}

	sqlDB := db.DB
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return &PDB{DB: db}, nil
}

func (db *PDB) Close() error {
	if db == nil || db.DB == nil {
		return nil
	}
	return db.DB.Close()
}

func (db *PDB) Migrate(path string) error {

	driver, err := postgres.WithInstance(db.DB.DB, &postgres.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("driver error(postgres): %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+path,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("migrate init error(postgres): %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration error(postgres): %w", err)
	}

	return nil
}
