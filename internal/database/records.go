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

type pgRecords struct {
	ModifiedTime pgtype.Timestamptz        `db:"modified_time"`
	Period       pgtype.Range[pgtype.Date] `db:"period"`
	SourceType   pgtype.Text               `db:"source_type"`
	Source       pgtype.Text               `db:"source"`
	Data         []byte                    `db:"data"`
	Compressed   pgtype.Bool               `db:"compressed"`
}

func AddRecords(ctx context.Context, records domain.Records) error {
	if len(records.Rows) < 1 {
		return fmt.Errorf("no rows, skipping")
	}
	semester, err := GetSemester(ctx, records.SemesterEnd)
	if err == pgx.ErrNoRows {
		semester, err = AddSemester(ctx, records.SemesterEnd)
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
		err = CompressJSON(&buf, records.Rows)
		if err != nil {
			return err
		}
		data = buf.Bytes()
	} else {
		// raw data bytes
		data, err = json.Marshal(records.Rows)
		if err != nil {
			return err
		}
	}

	return queryZero(ctx, `
		INSERT INTO records(modified_time, semester_id, source_id, data, compressed) VALUES
			($1, $2,
			(SELECT id FROM sources WHERE source_type = $3 AND LOWER(name) = LOWER($4)),
			$5, $6)
		ON CONFLICT (semester_id, source_id)
		DO UPDATE SET
			modified_time = $1,
			data = $5,
			compressed = $6
		`,
		pgtype.Timestamptz{Time: time.Unix(records.ModifiedTime, 0), Valid: true},
		semester.Id, records.SourceType, records.Source, data, default_compressed,
	)
}

func GetRecords(ctx context.Context, source_type, source string) (domain.Records, error) {
	var res domain.Records
	row, err := queryOne[pgRecords](ctx, `
		SELECT
		    r.modified_time,
		    sm.period,
		    s.source_type,
		    s.name AS source,
		    r.data,
			r.compressed
		FROM records r
		JOIN sources s ON r.source_id = s.id
		JOIN semesters sm ON r.semester_id = sm.id
		WHERE s.source_type = $1
		AND LOWER(s.name) = LOWER($2)
		ORDER BY r.modified_time DESC
		LIMIT 1;
	`, source_type, source)
	if err != nil {
		return res, err
	}

	res.ModifiedTime = row.ModifiedTime.Time.Unix()
	res.SourceType = row.SourceType.String
	res.Source = row.Source.String

	// decompress data if needed
	var rows []domain.RecordRow
	if row.Compressed.Bool {
		rows, err = DecompressJSON[[]domain.RecordRow](row.Data)
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

	res.SemesterStart = row.Period.Lower.Time.Format("02.01.2006")
	res.SemesterEnd = row.Period.Upper.Time.Add(-24 * time.Hour).Format("02.01.2006")

	return res, nil
}
