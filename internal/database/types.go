package database

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// tables

type SourceCategory struct {
	Id   pgtype.Int4
	Name pgtype.Text
}

//	type Source struct {
//		Id               pgtype.Int4
//		Name             pgtype.Text
//		SourceType       pgtype.Text
//		SourceCategoryId pgtype.Int4
//		Status           pgtype.Text
//	}
type LessonCategory struct {
	Id   pgtype.Int4
	Name pgtype.Text
}
type Lesson struct {
	Id               pgtype.Int4
	Name             pgtype.Text
	LessonCategoryId pgtype.Int4
}
type Semester struct {
	Id          pgtype.Int4
	Period      pgtype.Range[pgtype.Date]
	Description pgtype.Text
}
type Schedule struct {
	ModifiedTime pgtype.Timestamptz
	Period       pgtype.Range[pgtype.Date]
	SourceId     pgtype.Int4
	Days         pgtype.JSONBCodec
}
type Record struct {
	ModifiedTime pgtype.Timestamptz
	SemesterId   pgtype.Int4
	SourceId     pgtype.Int4
	Rows         pgtype.JSONBCodec
}
type Journal struct {
	Id           pgtype.Int4
	ModifiedTime pgtype.Timestamptz
	ModifiedDate pgtype.Date
	SemesterId   pgtype.Int4
	LessonId     pgtype.Int4
	SourceId     pgtype.Int4
	Rows         pgtype.JSONBCodec
}
