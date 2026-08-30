package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitPool(ctx context.Context, conn_uri string) error {
	config, err := pgxpool.ParseConfig(conn_uri)
	if err != nil {
		return err
	}
	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	DB, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}

	err = DB.Ping(ctx)
	if err != nil {
		return err
	}

	return queryZero(ctx, `
		DO $$ BEGIN

		CREATE TYPE source_type_enum AS ENUM ('group', 'lecturer', 'room');
		CREATE TYPE source_status AS ENUM ('active', 'inactive');

		CREATE TABLE source_categories (
		    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		    name VARCHAR(255) NOT NULL UNIQUE
		);

		CREATE TABLE sources (
		    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		    name VARCHAR(255) NOT NULL,
		    source_type source_type_enum NOT NULL,
		    source_category_id INT REFERENCES source_categories(id),
		    status source_status NOT NULL DEFAULT 'active',
		    CONSTRAINT uc_name_source_type UNIQUE (name, source_type)
		);

		CREATE TABLE lesson_categories (
		    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		    name VARCHAR(255) NOT NULL UNIQUE
		);

		CREATE TABLE lessons (
		    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		    name TEXT NOT NULL UNIQUE,
		    lesson_category_id INT REFERENCES lesson_categories(id)
		);

		CREATE TABLE semesters (
		    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		    period DATERANGE NOT NULL,
		    description TEXT,
		    EXCLUDE USING gist (period WITH &&)
		);


		------------ main tables ------------

		CREATE TABLE schedules (
		    modified_time TIMESTAMPTZ NOT NULL,
		    period DATERANGE NOT NULL,
		    source_id INT NOT NULL REFERENCES sources(id),
		    data BYTEA,
		    compressed BOOLEAN NOT NULL DEFAULT false,
		    PRIMARY KEY (modified_time, source_id)
		);

		CREATE TABLE records (
		    modified_time TIMESTAMPTZ NOT NULL,
		    semester_id INT NOT NULL REFERENCES semesters(id),
		    source_id INT NOT NULL REFERENCES sources(id),
		    data BYTEA,
		    compressed BOOLEAN NOT NULL DEFAULT false,
		    PRIMARY KEY (semester_id, source_id)
		);

		CREATE TABLE journals (
		    modified_time TIMESTAMPTZ NOT NULL,
		    modified_date DATE NOT NULL,
		    semester_id INT NOT NULL REFERENCES semesters(id),
		    journal_id INT NOT NULL, -- not primary key
		    source_id INT NOT NULL REFERENCES sources(id),
		    subgroup INT NOT NULL,
		    lesson_id INT NOT NULL REFERENCES lessons(id),
		    data BYTEA,
		    compressed BOOLEAN NOT NULL DEFAULT false,
		    CONSTRAINT uc_semester_source_subgroup_lesson UNIQUE (semester_id, source_id, subgroup, lesson_id)
		);

		EXCEPTION
			WHEN duplicate_object THEN NULL;
		END $$;
	`)
}
func ClosePool() {
	if DB != nil {
		DB.Close()
	}
}

func query[T any](ctx context.Context, q string, args ...any) ([]T, error) {
	rows, err := DB.Query(ctx, q, args...)
	if err != nil {
		return []T{}, err
	}

	rows_st, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return []T{}, err
	}

	return rows_st, nil
}

func queryOne[T any](ctx context.Context, q string, args ...any) (T, error) {
	var res T
	rows, err := DB.Query(ctx, q, args...)
	if err != nil {
		rows.Close()
		return res, err
	}

	rows_st, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	rows.Close()
	if err != nil {
		return res, err
	}

	if len(rows_st) < 1 {
		return res, pgx.ErrNoRows
	}

	res = rows_st[0]
	return res, nil
}

func queryZero(ctx context.Context, q string, args ...any) error {
	rows, err := DB.Query(ctx, q, args...)
	rows.Close()
	return err
}

func CheckModifiedTime(ctx context.Context, modified_time int64) (int64, error) {
	row, err := queryOne[struct {
		ModifiedTime pgtype.Timestamptz `db:"modified_time"`
	}](ctx, `SELECT modified_time FROM schedules WHERE modified_time = to_timestamp($1) LIMIT 1`,
		modified_time)
	return row.ModifiedTime.Time.Unix(), err
}
