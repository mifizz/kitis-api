package parser

import (
	"context"
	"slices"
	"strings"

	"kitis-api/internal/domain"
	api_errors "kitis-api/internal/errors"

	"github.com/PuerkitoBio/goquery"
)

func (p *Parser) parseRecords(doc *goquery.Document) (domain.Records, error) {
	var res domain.Records
	fixNewlines(doc)

	res.ModifiedTime = parseModifiedTime(doc)

	sources := source_regex.FindStringSubmatch(doc.Find("h1").Text())
	if len(sources) < 3 {
		return res, api_errors.ErrParserNotImplemented.WithMessage("Failed to parse source.")
	}
	res.SourceType = source_types_ru_en[strings.TrimSpace(sources[1])]
	res.Source = strings.TrimSpace(sources[2])

	table := doc.Find("table.inf > tbody > tr")
	if table.Length() < 1 {
		return res, api_errors.ErrParserNotImplemented.WithMessage("Table not found.")
	}

	table.Slice(1, table.Length()).Each(func(i int, s *goquery.Selection) {
		s_td := s.Find("td")
		if i == 0 {
			res.SemesterEnd = s_td.Eq(12).Text()
		}

		subgroup := stringToInt(s_td.Eq(3).Text())
		journal_url, _ := s_td.Eq(4).Find("a").Attr("href")
		journal_url, _ = strings.CutPrefix(journal_url, "j")
		journal_url, _ = strings.CutSuffix(journal_url, ".htm")
		load_semester := domain.RecordsLoadSemester{
			Total:     selToInt(s_td.Eq(6)),
			Planned:   selToInt(s_td.Eq(7)),
			Actual:    selToInt(s_td.Eq(8)),
			Remaining: selToInt(s_td.Eq(9)),
		}
		load_week := domain.RecordsLoadWeek{
			Planned: selToFloat(s_td.Eq(10), 1),
			Actual:  selToInt(s_td.Eq(11)),
		}
		progress_s, _ := s_td.Eq(13).Find("img").Attr("alt")
		row := domain.RecordRow{
			Lecturer:   s_td.Eq(1).Text(),
			Group:      s_td.Eq(2).Text(),
			Subgroup:   subgroup,
			Name:       s_td.Eq(4).Text(),
			JournalId:  stringToInt(journal_url),
			LessonType: s_td.Eq(5).Text(),
			Load: domain.RecordsLoad{
				Semester: load_semester,
				Week:     load_week,
			},
			Progress: stringToInt(progress_s),
		}
		res.Rows = append(res.Rows, row)
	})

	return res, nil
}

// --- public ---

func (p *Parser) GetRecords(ctx context.Context, source_type, source string) (domain.Records, error) {
	var res domain.Records
	if !slices.Contains(valid_source_types, source_type) {
		return res, api_errors.ErrParserSourceTypeInvalid
	}
	found_source_id, ok := p.FindSourceID(source_type, source)
	if !ok {
		return res, api_errors.ErrParserSourceNotFound
	}
	url := p.config.BaseURI + "/v" + source_type_abbrs[source_type] + found_source_id + p.config.URIsuffix
	doc, err := p.fetchDoc(ctx, url)
	if err != nil {
		return res, err
	}
	return p.parseRecords(doc)
}
