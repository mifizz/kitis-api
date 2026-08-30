package parser

import (
	"context"
	"kitis-api/internal/config"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	weekdays_short_long = map[string]string{
		"Пн":   "Понедельник",
		"Пн-1": "Понедельник",
		"Пн-2": "Понедельник",
		"Вт":   "Вторник",
		"Вт-1": "Вторник",
		"Вт-2": "Вторник",
		"Ср":   "Среда",
		"Ср-1": "Среда",
		"Ср-2": "Среда",
		"Чт":   "Четверг",
		"Чт-1": "Четверг",
		"Чт-2": "Четверг",
		"Пт":   "Пятница",
		"Пт-1": "Пятница",
		"Пт-2": "Пятница",
		"Сб":   "Суббота",
		"Сб-1": "Суббота",
		"Сб-2": "Суббота",
		"Вс":   "Воскресенье",
		"Вс-1": "Воскресенье",
		"Вс-2": "Воскресенье",
	}
	// weekdays_short_long = map[string]string{
	// 	"Пн": "Понедельник",
	// 	"Вт": "Вторник",
	// 	"Ср": "Среда",
	// 	"Чт": "Четверг",
	// 	"Пт": "Пятница",
	// 	"Сб": "Суббота",
	// 	"Вс": "Воскресенье",
	// }
	source_types_ru_en = map[string]string{
		"Группа":        "group",
		"Преподаватель": "lecturer",
		"Аудитория":     "room",
	}
	source_regex       = regexp.MustCompile(`([^:]+): ([^$]+)`)
	date_regex         = regexp.MustCompile(`\d{2}\.\d{2}\.\d{4}`)
	lesson_fix_regex   = regexp.MustCompile(`\s*\(\d+\)`)
	valid_source_types = []string{"group", "lecturer", "room"}
	source_type_abbrs  = map[string]string{"group": "g", "lecturer": "p", "room": "a"}
)

type Parser struct {
	client  *http.Client
	config  *config.ParserConfig
	sources map[string]map[string]string
}

func NewParser(client *http.Client, conf *config.ParserConfig) *Parser {
	return &Parser{client: client, config: conf}
}

func stringToInt(s string) int {
	number, _ := strconv.Atoi(s)
	return number
}

// modified time parsers

func parseModifiedTimeString(mt string) int64 {
	loc, _ := time.LoadLocation("Europe/Kaliningrad")
	t, err := time.ParseInLocation("Обновлено: 02.01.2006 в 15:04.", mt, loc)
	if err != nil {
		return 0
	}
	return t.Unix()
}

func (p *Parser) GetModifiedTime(ctx context.Context) (int64, error) {
	url := p.config.BaseURI + p.config.URIprefixes["group"] + p.config.URIsuffix
	doc, err := p.fetchDoc(ctx, url)
	if err != nil {
		return 0, err
	}
	return parseModifiedTime(doc), nil
}
