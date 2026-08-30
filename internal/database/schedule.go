package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kitis-api/internal/domain"
	"time"
	_ "time/tzdata"

	"github.com/jackc/pgx/v5/pgtype"
)

type pgSchedule struct {
	ModifiedTime pgtype.Timestamptz `db:"modified_time"`
	// Period       pgtype.Range[pgtype.Date] `db:"period"`
	SourceType pgtype.Text `db:"source_type"`
	Source     pgtype.Text `db:"source"`
	Data       []byte      `db:"data"`
	Compressed pgtype.Bool `db:"compressed"`
}
type pgScheduleVersion struct {
	Timestamp pgtype.Timestamptz        `db:"timestamp"`
	Period    pgtype.Range[pgtype.Date] `db:"period"`
}

func AddSchedule(ctx context.Context, schedule domain.Schedule) error {
	if len(schedule.Days) < 7 {
		return fmt.Errorf("schedule week is not complete, expected 7 days, got %d", len(schedule.Days))
	}
	period_start, err := time.Parse("02.01.2006", schedule.Days[0].Date)
	if err != nil {
		return fmt.Errorf("failed to parse schedule period start: %v", err)
	}
	period_end, err := time.Parse("02.01.2006", schedule.Days[len(schedule.Days)-1].Date)
	if err != nil {
		return fmt.Errorf("failed to parse schedule period end: %v", err)
	}

	const default_compressed bool = true
	var data []byte

	if default_compressed {
		// compress data
		var buf bytes.Buffer
		err = CompressJSON(&buf, schedule.Days)
		if err != nil {
			return err
		}
		data = buf.Bytes()
	} else {
		// raw data bytes
		data, err = json.Marshal(schedule.Days)
		if err != nil {
			return err
		}
	}

	return queryZero(ctx, `
		INSERT INTO schedules(modified_time, period, source_id, data, compressed) VALUES
			($1, $2,
			(SELECT id FROM sources WHERE source_type = $3 AND LOWER(name) = LOWER($4)),
			$5, $6)
		ON CONFLICT (modified_time, source_id)
		DO UPDATE SET
			period = $2,
			data = $5,
			compressed = $6
		`,
		pgtype.Timestamptz{Time: time.Unix(schedule.ModifiedTime, 0), Valid: true},
		pgtype.Range[pgtype.Date]{
			Lower:     pgtype.Date{Time: period_start, Valid: true},
			Upper:     pgtype.Date{Time: period_end, Valid: true},
			LowerType: pgtype.Inclusive,
			UpperType: pgtype.Inclusive,
			Valid:     true,
		},
		schedule.SourceType, schedule.Source, data, default_compressed)
}

func GetScheduleNewest(ctx context.Context, source_type, source string) (domain.Schedule, error) {
	var res domain.Schedule
	row, err := queryOne[pgSchedule](ctx, `
		SELECT
		    schd.modified_time,
		    s.source_type,
		    s.name AS source,
		    schd.data,
			schd.compressed
		FROM schedules schd
		JOIN sources s ON schd.source_id = s.id
		WHERE s.source_type = $1
		AND LOWER(s.name) = LOWER($2)
		ORDER BY schd.modified_time DESC
		LIMIT 1
	`, source_type, source)
	if err != nil {
		return res, err
	}

	res.ModifiedTime = row.ModifiedTime.Time.Unix()
	res.SourceType = row.SourceType.String
	res.Source = row.Source.String

	// decompress data if needed
	var days []domain.ScheduleDay
	if row.Compressed.Bool {
		days, err = DecompressJSON[[]domain.ScheduleDay](row.Data)
		if err != nil {
			return res, err
		}
	} else {
		err := json.Unmarshal(row.Data, &days)
		if err != nil {
			return res, err
		}
	}
	res.Days = days

	return res, nil
}

func GetScheduleByTimestamp(ctx context.Context, source_type, source string, timestamp int64) (domain.Schedule, error) {
	row, err := queryOne[pgSchedule](ctx, `
		SELECT
		    schd.modified_time,
		    s.source_type,
		    s.name AS source,
		    schd.data,
			schd.compressed
		FROM schedules schd
		JOIN sources s ON schd.source_id = s.id
		WHERE s.source_type = $1
		AND LOWER(s.name) = LOWER($2)
		AND schd.modified_time = to_timestamp($3)
		LIMIT 1
	`, source_type, source, timestamp)
	if err != nil {
		return domain.Schedule{}, err
	}

	var res domain.Schedule
	res.ModifiedTime = row.ModifiedTime.Time.Unix()
	res.SourceType = row.SourceType.String
	res.Source = row.Source.String

	// decompress data if needed
	var days []domain.ScheduleDay
	if row.Compressed.Bool {
		days, err = DecompressJSON[[]domain.ScheduleDay](row.Data)
		if err != nil {
			return res, err
		}
		res.Days = days
	} else {
		err := json.Unmarshal(row.Data, &days)
		if err != nil {
			return res, err
		}
	}
	res.Days = days

	return res, nil
}

func GetScheduleVersions(ctx context.Context, source_type, source string) ([]domain.ScheduleVersion, error) {
	rows, err := query[pgScheduleVersion](ctx, `
		SELECT
		    schd.modified_time AS timestamp,
		    schd.period
		FROM schedules schd
		JOIN sources s ON schd.source_id = s.id
		WHERE s.source_type = $1
		AND LOWER(s.name) = LOWER($2)
	`, source_type, source)
	if err != nil {
		return []domain.ScheduleVersion{}, err
	}

	loc, err := time.LoadLocation("Europe/Kaliningrad")
	if err != nil {
		return []domain.ScheduleVersion{}, err
	}

	var res []domain.ScheduleVersion
	for _, row := range rows {
		t := time.Unix(row.Timestamp.Time.Unix(), 0)
		lt := t.In(loc)

		human := lt.Format("02.01.2006 15:04:05-0700")
		period_start := row.Period.Lower.Time.Format("02.01.2006")
		period_end := row.Period.Upper.Time.Add(-24 * time.Hour).Format("02.01.2006")

		res = append(res, domain.ScheduleVersion{
			Timestamp:   row.Timestamp.Time.Unix(),
			Human:       human,
			PeriodStart: period_start,
			PeriodEnd:   period_end,
		})
	}
	return res, nil
}

func GetScheduleByIDNewest(ctx context.Context, id int) (domain.Schedule, error) {
	var res domain.Schedule
	row, err := queryOne[pgSchedule](ctx, `
		SELECT
		    schd.modified_time,
		    s.source_type,
		    s.name AS source,
		    schd.data,
			schd.compressed
		FROM schedules schd
		JOIN sources s ON schd.source_id = s.id
		WHERE s.id = $1
		ORDER BY schd.modified_time DESC
		LIMIT 1
	`, id)
	if err != nil {
		return res, err
	}

	res.ModifiedTime = row.ModifiedTime.Time.Unix()
	res.SourceType = row.SourceType.String
	res.Source = row.Source.String

	// decompress data if needed
	var days []domain.ScheduleDay
	if row.Compressed.Bool {
		days, err = DecompressJSON[[]domain.ScheduleDay](row.Data)
		if err != nil {
			return res, err
		}
	} else {
		err := json.Unmarshal(row.Data, &days)
		if err != nil {
			return res, err
		}
	}
	res.Days = days

	return res, nil
}
