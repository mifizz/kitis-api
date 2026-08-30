package domain

type Journal struct {
	ModifiedTime  int64        `json:"modified_time"`
	ModifiedDate  string       `json:"modified_date"`
	SemesterStart string       `json:"semester_start"`
	SemesterEnd   string       `json:"semester_end"`
	JournalId     int          `json:"journal_id"`
	Group         string       `json:"group"`
	Subgroup      int          `json:"subgroup"`
	Lesson        string       `json:"lesson"`
	Rows          []JournalRow `json:"rows"`
}
type JournalRow struct {
	Date         string `json:"date"`
	LessonNumber int    `json:"lesson_number"`
	Lecturer     string `json:"lecturer"`
	Room         string `json:"room"`
}
