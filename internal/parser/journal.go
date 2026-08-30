package parser

import (
	"context"
	"kitis-api/internal/database"
	"kitis-api/internal/domain"
	api_errors "kitis-api/internal/errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jackc/pgx/v5"
)

func (p *Parser) parseJournal(doc *goquery.Document) (domain.Journal, error) {
	fixNewlines(doc)
	var res domain.Journal

	res.ModifiedTime = parseModifiedTime(doc)
	res.ModifiedDate = strings.TrimSpace(doc.Find("li.zgr").Next().Text())
	res.Lesson = strings.TrimSpace(doc.Find("li.zgr").Text())
	if res.Lesson == "" {
		return res, api_errors.ErrParserJournalNotFound
	}
	semester, err := database.GetSemester(context.Background(), res.ModifiedDate)
	if err == pgx.ErrNoRows {
		res.SemesterStart = ""
		res.SemesterEnd = ""
	} else if err != nil {
		return res, err
	} else {
		res.SemesterStart = semester.Period.Lower.Time.Format("02.01.2006")
		res.SemesterEnd = semester.Period.Upper.Time.Add(-24 * time.Hour).Format("02.01.2006")
	}

	// Переменная для хранения суммы подгрупп.
	// Это нужно для высчитывания средней подгруппы на случай,
	// если в некоторых строках журнала была ошибочно подписана не та подгруппа
	subgroup_sum := 0

	table := doc.Find("table.inf > tbody > tr")
	if table.Length() < 1 {
		return res, api_errors.ErrParserJournalNotFound
	}

	table.Slice(1, table.Length()).Each(func(i int, s *goquery.Selection) {
		s_td := s.Find("td")
		subgroup_sum += selToInt(s_td.Eq(4))

		row := domain.JournalRow{
			Date:         s_td.Eq(1).Text(),
			LessonNumber: selToInt(s_td.Eq(2)),
			Lecturer:     s_td.Eq(5).Text(),
			Room:         s_td.Eq(6).Text(),
		}

		// Последняя строка журнала
		if i == table.Length()-2 {
			res.Group = s_td.Eq(3).Text()
			res.Subgroup = int(math.Round(float64(subgroup_sum) / float64(table.Length()-1)))
		}

		res.Rows = append(res.Rows, row)
	})

	return res, nil
}

// --- public ---

func (p *Parser) GetJournal(ctx context.Context, journal_id int) (domain.Journal, error) {
	var res domain.Journal
	url := p.config.BaseURI + "/j" + strconv.Itoa(journal_id) + p.config.URIsuffix
	doc, err := p.fetchDoc(ctx, url)
	if err != nil {
		return res, err
	}

	res, err = p.parseJournal(doc)
	if err != nil {
		return res, err
	}
	res.JournalId = journal_id
	return res, nil
}
