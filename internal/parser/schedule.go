package parser

import (
	"context"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"kitis-api/internal/domain"
	api_errors "kitis-api/internal/errors"

	"github.com/PuerkitoBio/goquery"
)

// если ты читаешь этот комментарий, то фан-факт:
// я хз, зачем я тут писал комментарии на русском, если во всём остальном проекте они на английском :man_shrugging:
// и нет, этот код я писал не через нейронку, просто я ЗАЧЕМ-ТО решил подписать комментарии
// возможно, чтобы потом самим разобраться, если буду фиксить/переписывать, а то этот китис расписание просто комшар

func (p *Parser) parseSchedule(doc *goquery.Document) (domain.Schedule, error) {
	var res domain.Schedule
	// Замена <br> на \\n для предотвращения проблем с selection.Text()
	fixNewlines(doc)

	res.ModifiedTime = parseModifiedTime(doc)

	sources := source_regex.FindStringSubmatch(doc.Find("h1").Text())
	if len(sources) < 3 {
		return res, api_errors.ErrParserNotImplemented.WithMessage("Failed to parse source.")
	}
	res.SourceType = source_types_ru_en[strings.TrimSpace(sources[1])]
	res.Source = strings.TrimSpace(sources[2])

	// Разговоры о важном между 5 и 6 парами
	lesson_rov_second := domain.ScheduleLesson{
		Number: -1,
		Bells:  p.config.Bells.Monday[-1],
	}

	table := doc.Find("table.inf > tbody > tr")
	if table.Length() < 4 {
		return res, api_errors.ErrParserNotImplemented.WithMessage("Table not found.")
	}
	day := domain.ScheduleDay{}

	table.Slice(2, table.Length()).Each(func(i int, s *goquery.Selection) {
		var (
			start_idx int                = 0
			sel_td    *goquery.Selection = s.Find("td")
			text0     string             = strings.TrimSpace(sel_td.First().Text())
		)
		// Проверяем наличие даты в первой ячейке строки
		if (!strings.Contains(text0, ":") || date_regex.MatchString(text0)) && len(text0) > 1 {
			// Индекс парсинга строки = 1
			start_idx = 1
			var (
				text0_split = strings.Split(text0, "\\n")
				date        = text0_split[0]
				weekday     = text0_split[1]
			)
			// Проверяем, был ли уже заполнен day
			if len(day.Date) > 0 {
				// Замена первых разговоров о важном на вторые, если есть 5 и 6 пары
				if day.Weekday == "Понедельник" && len(day.Lessons) > 0 {
					lesson2_idx, lesson6_idx := -1, -1
					rov_exists := false
					// Проходимся по
					for i, lesson := range day.Lessons {
						if strings.Contains(strings.ToLower(lesson.Name), "разговоры о важном") {
							rov_exists = true
						}
						switch lesson.Number {
						case 2:
							lesson2_idx = i
						case 6:
							lesson6_idx = i
						}
					}
					// Пересоздаём пару РоВ на основе указанной в расписании
					// update 13.11.2026: Почему-то у ИМС25-1 нет первого РоВа в понедельник (с учётом, что пары с 4 по 6), а у других групп - есть
					// следовательно, пропускаем всю эту канитель, если РоВа нет вообще
					if rov_exists {
						lesson_rov := lesson_rov_second
						lesson_rov.Name = day.Lessons[0].Name
						lesson_rov.Group = day.Lessons[0].Group
						lesson_rov.Lecturer = day.Lessons[0].Lecturer
						lesson_rov.Room = day.Lessons[0].Room
						if lesson2_idx == -1 && lesson6_idx > -1 {
							// Вставляем второй РоВ
							day.Lessons = append(day.Lessons[:lesson6_idx], append([]domain.ScheduleLesson{lesson_rov}, day.Lessons[lesson6_idx:]...)...)
						}
						if lesson2_idx == -1 {
							// Убираем первый РоВ
							day.Lessons = day.Lessons[1:]
						}
					}
				}
				// Добавляем заполненный день в список
				res.Days = append(res.Days, day)
			}
			// Переделываем сокращение дня в полную версию
			weekday_long := weekdays_short_long[weekday]
			// ... или нет
			if weekday_long == "" {
				weekday_long = weekday
			}
			// Обнуляем день и записываем в него текущие значения даты и дня недели
			day = domain.ScheduleDay{Date: date, Weekday: weekday_long, Lessons: []domain.ScheduleLesson{}}
		} else if sel_td.Length() >= 2 {
			// Индекс парсинга = 0. Есть хотя бы одна пара
			start_idx = 0
		} else {
			// Нет ни даты, ни пар. Пропускаем строку
			return
		}
		if sel_td.Length() < start_idx+2 {
			// Пропускаем пустую строку (разделитель между днями)
			return
		}

		// Достаём номер пары
		// Для этого переводим текст элемента в utf-8, потом в строку и потом в число, я уже не помню честно, почему я так написал
		lesson_n, err := strconv.Atoi(string([]rune(sel_td.Eq(start_idx).Text())[0]))
		if err != nil {
			slog.Error("Lesson number parse error", "error", err)
			lesson_n = 0
		}

		sel_td.Slice(start_idx+1, sel_td.Length()).Each(func(i int, s *goquery.Selection) {
			// Пропускаем, если в элементе нет текста
			if len(strings.TrimSpace(s.Text())) == 0 {
				return
			}
			zoo := make(map[string]string)
			for _, z := range []string{"z1", "z2", "z3"} {
				zoo[z] = strings.Join(s.Children().Filter("."+z).Map(func(i int, s *goquery.Selection) string {
					text, _ := s.Html()
					return strings.TrimSpace(text)
				}), ", ")
			}
			z_mapping := map[string]map[string]string{
				"group":    {"name": zoo["z1"], "group": res.Source, "room": zoo["z2"], "lecturer": zoo["z3"]},
				"lecturer": {"name": zoo["z3"], "lecturer": res.Source, "room": zoo["z2"], "group": zoo["z1"]},
				"room":     {"name": zoo["z3"], "room": res.Source, "group": zoo["z2"], "lecturer": zoo["z1"]},
			}
			subgroup := 0
			if sel_td.Length() > start_idx+2 {
				subgroup = i + 1
			}
			bells := p.config.Bells.General[lesson_n]
			if day.Weekday == "Понедельник" {
				bells = p.config.Bells.Monday[lesson_n]
			}

			lesson := domain.ScheduleLesson{
				Number:   lesson_n,
				Bells:    bells,
				Subgroup: subgroup,
				Name:     z_mapping[res.SourceType]["name"],
				Group:    z_mapping[res.SourceType]["group"],
				Lecturer: z_mapping[res.SourceType]["lecturer"],
				Room:     z_mapping[res.SourceType]["room"],
			}
			if len(lesson.Name) > 0 {
				lesson.Name = lesson_fix_regex.ReplaceAllString(lesson.Name, "")
			}
			if len(day.Date) > 0 {
				day.Lessons = append(day.Lessons, lesson)
			}
		})
	})
	if len(day.Date) > 0 {
		res.Days = append(res.Days, day)
	}
	return res, nil
}

// --- public ---

func (p *Parser) GetSchedule(ctx context.Context, source_type, source string) (domain.Schedule, error) {
	var res domain.Schedule
	if !slices.Contains(valid_source_types, source_type) {
		return res, api_errors.ErrParserSourceTypeInvalid
	}
	found_source_id, ok := p.FindSourceID(source_type, source)
	if !ok {
		return res, api_errors.ErrParserSourceNotFound
	}
	url := p.config.BaseURI + "/c" + source_type_abbrs[source_type] + found_source_id + p.config.URIsuffix
	doc, err := p.fetchDoc(ctx, url)
	if err != nil {
		return res, err
	}
	return p.parseSchedule(doc)
}
