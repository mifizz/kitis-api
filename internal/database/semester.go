package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type pgSemester struct {
	Id          pgtype.Int4               `db:"id"`
	Period      pgtype.Range[pgtype.Date] `db:"period"`
	Description pgtype.Text               `db:"description"`
}

func AddSemester(ctx context.Context, semester_date_s string) (pgSemester, error) {
	semester_date, err := time.Parse("02.01.2006", semester_date_s)
	if err != nil {
		return pgSemester{}, fmt.Errorf("failed to parse semester date: '%s'", semester_date_s)
	}

	var semester_start time.Time
	var semester_end time.Time
	if year, month, _ := semester_date.Date(); month >= time.January && month <= time.June {
		// semester 1
		semester_start = time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		semester_end = time.Date(year, time.June, 30, 0, 0, 0, 0, time.UTC)
	} else if year, month, _ := semester_date.Date(); month >= time.September && month <= time.December {
		// semester 2
		semester_start = time.Date(year, time.September, 1, 0, 0, 0, 0, time.UTC)
		semester_end = time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	} else {
		// bad semester date
		return pgSemester{}, fmt.Errorf("bad semester date: %v", semester_date.Format("02.01.2006"))
	}
	period := pgtype.Range[pgtype.Date]{
		Lower:     pgtype.Date{Time: semester_start, Valid: true},
		Upper:     pgtype.Date{Time: semester_end, Valid: true},
		LowerType: pgtype.Inclusive,
		UpperType: pgtype.Inclusive,
		Valid:     true,
	}
	return queryOne[pgSemester](ctx, `INSERT INTO semesters(period) VALUES ($1) RETURNING id, period, description`, period)
}

func GetSemester(ctx context.Context, semester_date_s string) (pgSemester, error) {
	semester_date, err := time.Parse("02.01.2006", semester_date_s)
	if err != nil {
		return pgSemester{}, fmt.Errorf("failed to parse semester date: '%s'", semester_date_s)
	}
	return queryOne[pgSemester](ctx, `SELECT id, period, description FROM semesters WHERE period @> $1::date`, pgtype.Date{Time: semester_date, Valid: true})
}
