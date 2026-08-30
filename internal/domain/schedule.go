package domain

type Schedule struct {
	ModifiedTime int64         `json:"modified_time"`
	SourceType   string        `json:"source_type"` // group | lecturer | room
	Source       string        `json:"source"`
	Days         []ScheduleDay `json:"days"`
}
type ScheduleDay struct {
	Date    string           `json:"date"`
	Weekday string           `json:"weekday"`
	Lessons []ScheduleLesson `json:"lessons"`
}
type ScheduleLesson struct {
	Number   int    `json:"number"`
	Bells    string `json:"bells"`
	Subgroup int    `json:"subgroup"`
	Name     string `json:"name"`
	Group    string `json:"group"`
	Lecturer string `json:"lecturer"`
	Room     string `json:"room"`
}

type ScheduleVersion struct {
	Timestamp   int64  `json:"timestamp"`
	Human       string `json:"human"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}
