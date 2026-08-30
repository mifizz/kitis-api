package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kitis-api/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type pgLesson struct {
	Id               pgtype.Int4 `db:"id"`
	Name             pgtype.Text `db:"name"`
	LessonCategoryId pgtype.Int4 `db:"lesson_category_id"`
}
type pgJournal struct {
	ModifiedTime pgtype.Timestamptz        `db:"modified_time"`
	ModifiedDate pgtype.Date               `db:"modified_date"`
	Period       pgtype.Range[pgtype.Date] `db:"period"`
	JournalId    pgtype.Int4               `db:"journal_id"`
	Group        pgtype.Text               `db:"group"`
	Subgroup     pgtype.Int4               `db:"subgroup"`
	Lesson       pgtype.Text               `db:"lesson"`
	Data         []byte                    `db:"data"`
	Compressed   pgtype.Bool               `db:"compressed"`
}

func AddLesson(ctx context.Context, name string) (pgLesson, error) {
	return queryOne[pgLesson](ctx, `
		INSERT INTO lessons (name) VALUES ($1)
		ON CONFLICT (name)
		DO UPDATE SET
			name = $1
		RETURNING id, name, lesson_category_id
	`, name)
}

func GetLesson(ctx context.Context, name string) (pgLesson, error) {
	return queryOne[pgLesson](ctx, `
		SELECT id, name, lesson_category_id FROM lessons
		WHERE LOWER(name) = LOWER($1)
		LIMIT 1
	`, name)
}

func AddJournal(ctx context.Context, journal domain.Journal) error {
	semester, err := GetSemester(ctx, journal.ModifiedDate)
	if err == pgx.ErrNoRows {
		semester, err = AddSemester(ctx, journal.ModifiedDate)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	modified_date, err := time.Parse("02.01.2006", journal.ModifiedDate)
	if err != nil {
		return fmt.Errorf("failed to parse modified date: %v", err)
	}

	lesson, err := GetLesson(ctx, journal.Lesson)
	if err == pgx.ErrNoRows {
		lesson, err = AddLesson(ctx, journal.Lesson)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	const default_compressed bool = true
	var data []byte

	if default_compressed {
		// compress data
		var buf bytes.Buffer
		err = CompressJSON(&buf, journal.Rows)
		if err != nil {
			return err
		}
		data = buf.Bytes()
	} else {
		// raw data bytes
		data, err = json.Marshal(journal.Rows)
		if err != nil {
			return err
		}
	}

	return queryZero(ctx, `
		INSERT INTO journals(modified_time, modified_date, semester_id, journal_id, source_id, subgroup, lesson_id, data, compressed) VALUES
		(
			$1,			-- modified time
			$2::date,	-- modified date
			$3,			-- semester id
			$4,			-- journal id
			(SELECT id FROM sources WHERE source_type = 'group' AND LOWER(name) = LOWER($5)), -- source id
			$6,			-- subgroup
			$7,			-- lesson id
			$8,			-- data
			$9			-- compressed
		)
		ON CONFLICT (semester_id, source_id, subgroup, lesson_id)
		DO UPDATE SET
		    modified_time = $1,
		    modified_date = $2::date,
		    journal_id = $4,
		    data = $8,
			compressed = $9
		`,
		pgtype.Timestamptz{Time: time.Unix(journal.ModifiedTime, 0), Valid: true},
		pgtype.Date{Time: modified_date, Valid: true},
		semester.Id, journal.JournalId, journal.Group, journal.Subgroup, lesson.Id, data, default_compressed,
	)
}

func GetJournal(ctx context.Context, source string, subgroup int, lesson string) (domain.Journal, error) {
	var res domain.Journal
	row, err := queryOne[pgJournal](ctx, `
		SELECT
		    j.modified_time,
		    j.modified_date,
		    sm.period,
		    j.journal_id,
		    s.name AS group,
		    j.subgroup,
		    l.name AS lesson,
		    j.data,
			j.compressed
		FROM journals j
		JOIN sources s ON j.source_id = s.id
		JOIN lessons l ON j.lesson_id = l.id
		JOIN semesters sm ON j.semester_id = sm.id
		WHERE s.source_type = 'group'
		AND LOWER(s.name) = LOWER($1)
		AND j.subgroup = $2
		AND LOWER(l.name) = LOWER($3)
		ORDER BY j.modified_time DESC
		LIMIT 1
	`, source, subgroup, lesson)
	if err != nil {
		return res, err
	}

	res.ModifiedTime = row.ModifiedTime.Time.Unix()
	res.ModifiedDate = row.ModifiedDate.Time.Format("02.01.2006")
	res.SemesterStart = row.Period.Lower.Time.Format("02.01.2006")
	res.SemesterEnd = row.Period.Upper.Time.Add(-24 * time.Hour).Format("02.01.2006")
	res.JournalId = int(row.JournalId.Int32)
	res.Group = row.Group.String
	res.Subgroup = int(row.Subgroup.Int32)
	res.Lesson = row.Lesson.String

	// decompress data if needed
	var rows []domain.JournalRow
	if row.Compressed.Bool {
		rows, err = DecompressJSON[[]domain.JournalRow](row.Data)
		if err != nil {
			return res, err
		}
	} else {
		err := json.Unmarshal(row.Data, &rows)
		if err != nil {
			return res, err
		}
	}
	res.Rows = rows

	return res, nil
}
