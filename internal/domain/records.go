package domain

type Records struct {
	ModifiedTime  int64       `json:"modified_time"`
	SemesterStart string      `json:"semester_start"`
	SemesterEnd   string      `json:"semester_end"`
	SourceType    string      `json:"source_type"` // group | lecturer | room
	Source        string      `json:"source"`
	Rows          []RecordRow `json:"rows"`
}
type RecordRow struct {
	Lecturer   string      `json:"lecturer"`
	Group      string      `json:"group"`
	Subgroup   int         `json:"subgroup"`
	Name       string      `json:"name"`
	JournalId  int         `json:"journal_id"`
	LessonType string      `json:"lesson_type"`
	Load       RecordsLoad `json:"load"`
	Progress   int         `json:"progress"`
}
type RecordsLoad struct {
	Semester RecordsLoadSemester `json:"semester"`
	Week     RecordsLoadWeek     `json:"week"`
}
type RecordsLoadSemester struct {
	Total     int `json:"total"`
	Planned   int `json:"planned"`
	Actual    int `json:"actual"`
	Remaining int `json:"remaining"`
}
type RecordsLoadWeek struct {
	Planned float64 `json:"planned"`
	Actual  int     `json:"actual"`
}
