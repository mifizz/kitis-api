package database

import (
	"context"
	"kitis-api/internal/domain"

	"github.com/jackc/pgx/v5/pgtype"
)

type dbSource struct {
	Id             int
	Name           string
	SourceType     string
	SourceCategory string
	Status         string
}

func AddSource(ctx context.Context, source_type, source string) error {
	return queryZero(ctx, `
		INSERT INTO sources (name, source_type) VALUES ($2, $1)
		ON CONFLICT (name, source_type)
		DO NOTHING
	`, source_type, source)
}

func SetSourceStatus(ctx context.Context, source_type, source, status string) error {
	return queryZero(ctx, `
		UPDATE sources
		SET status = $3
		WHERE source_type = $1 AND LOWER(name) = LOWER($2)
	`, source_type, source, status)
}

func GetSourcesActive(ctx context.Context) (map[string][]string, error) {
	res := map[string][]string{}
	rows, err := query[struct {
		SourceType pgtype.Text `db:"source_type"`
		Source     pgtype.Text `db:"source"`
	}](ctx, `
		SELECT DISTINCT
			source_type,
			name AS source
		FROM sources
		WHERE status = 'active'
		ORDER BY source_type, name
	`)
	if err != nil {
		return res, err
	}

	for _, row := range rows {
		if !row.SourceType.Valid || !row.Source.Valid {
			continue
		}
		res[row.SourceType.String] = append(res[row.SourceType.String], row.Source.String)
	}

	return res, nil
}
func GetSourcesExtended(ctx context.Context) (map[string]map[string]domain.Source, error) {
	res := map[string]map[string]domain.Source{}
	rows, err := query[struct {
		Id         pgtype.Int4 `db:"id"`
		SourceType pgtype.Text `db:"source_type"`
		Source     pgtype.Text `db:"source"`
		Category   pgtype.Text `db:"category"`
		Status     pgtype.Text `db:"status"`
	}](ctx, `
		SELECT DISTINCT
			s.id,
		    s.source_type,
		    s.name AS source,
		    sc.name AS category,
			s.status
		FROM sources s
		LEFT JOIN source_categories sc ON s.source_category_id = sc.id
		ORDER BY s.source_type, s.name
	`)
	if err != nil {
		return res, err
	}

	for _, row := range rows {
		if !row.SourceType.Valid || !row.Source.Valid {
			continue
		}
		if _, ok := res[row.SourceType.String]; !ok {
			res[row.SourceType.String] = map[string]domain.Source{}
		}
		res[row.SourceType.String][row.Source.String] = domain.Source{
			Id:       int(row.Id.Int32),
			Category: row.Category.String,
			Status:   row.Status.String,
		}
	}

	return res, nil
}

func GetSources(ctx context.Context, status string) (map[string][]string, error) {
	if status != "active" && status != "inactive" {
		status = "active"
	}
	rows, err := query[struct {
		SourceType pgtype.Text `db:"source_type"`
		Source     pgtype.Text `db:"source"`
	}](ctx, `
		SELECT DISTINCT
			source_type,
			name AS source
		FROM sources
		WHERE status = $1
		ORDER BY source_type, name
	`, status)
	if err != nil {
		return map[string][]string{}, err
	}

	res := map[string][]string{}
	for _, row := range rows {
		if !row.SourceType.Valid || !row.Source.Valid {
			continue
		}
		res[row.SourceType.String] = append(res[row.SourceType.String], row.Source.String)
	}

	return res, nil
}

func GetSource(ctx context.Context, source_type, source string) (dbSource, error) {
	var res dbSource
	row, err := queryOne[struct {
		Id         pgtype.Int4 `db:"id"`
		Name       pgtype.Text `db:"name"`
		SourceType pgtype.Text `db:"source_type"`
		Category   pgtype.Text `db:"category"`
		Status     pgtype.Text `db:"status"`
	}](ctx, `
		SELECT
			s.id,
			s.name,
			s.source_type,
			sc.name AS category,
			s.status
		FROM sources s
		LEFT JOIN source_categories sc ON s.source_category_id = sc.id
		WHERE s.source_type = $1
		AND LOWER(s.name) = LOWER($2)
		LIMIT 1
	`, source_type, source)
	if err != nil {
		return res, err
	}

	res = dbSource{
		Id:             int(row.Id.Int32),
		Name:           row.Name.String,
		SourceType:     row.SourceType.String,
		SourceCategory: row.Category.String,
		Status:         row.Status.String,
	}
	return res, err
}
func GetSourceByID(ctx context.Context, id int) (dbSource, error) {
	var res dbSource
	row, err := queryOne[struct {
		Id         pgtype.Int4 `db:"id"`
		Name       pgtype.Text `db:"name"`
		SourceType pgtype.Text `db:"source_type"`
		Category   pgtype.Text `db:"category"`
		Status     pgtype.Text `db:"status"`
	}](ctx, `
		SELECT
			s.id,
			s.name,
			s.source_type,
			sc.name AS category,
			s.status
		FROM sources s
		LEFT JOIN source_categories sc ON s.source_category_id = sc.id
		WHERE s.id = $1
		LIMIT 1
	`, id)
	if err != nil {
		return res, err
	}

	res = dbSource{
		Id:             int(row.Id.Int32),
		Name:           row.Name.String,
		SourceType:     row.SourceType.String,
		SourceCategory: row.Category.String,
		Status:         row.Status.String,
	}
	return res, err
}
