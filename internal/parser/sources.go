package parser

import (
	"context"
	"kitis-api/internal/database"
	"log/slog"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var source_id_regex = regexp.MustCompile(`(?:c|v)(?:g|p|a)(\d+)\.htm`)

func (p *Parser) parseDocSources(doc *goquery.Document) map[string]string {
	result := map[string]string{}
	doc.Find("a.z0").Each(func(i int, s *goquery.Selection) {
		source, _ := s.Html()
		uri, _ := s.Attr("href")
		match := source_id_regex.FindStringSubmatch(uri)
		if len(match) < 2 {
			return
		}
		id := match[1]
		result[source] = id
	})
	return result
}
func (p *Parser) parseAllSources(ctx context.Context) (map[string]map[string]string, error) {
	/* Структура объекта результата:
	 *
	 * group:
	 *   ИСп24-1: 102
	 * 	 ...
	 * lecturer:
	 *   Шевченко П.С.: 126
	 * 	 ...
	 * room:
	 *   "203": 104
	 * 	 ...
	 *
	 */
	result := map[string]map[string]string{}
	for _, page := range []string{"group", "lecturer", "room"} {
		doc, err := p.fetchDoc(ctx, p.config.BaseURI+p.config.URIprefixes[page]+p.config.URIsuffix)
		if err != nil {
			slog.Error("can't parse sources", "type", page, "error", err)
			return result, err
		}
		result[page] = p.parseDocSources(doc)
	}
	return result, nil
}
func (p *Parser) FindSourceID(source_type, source_in string) (string, bool) {
	for source, id := range p.sources[source_type] {
		if strings.EqualFold(source_in, source) {
			return id, true
		}
	}
	return "", false
}

// --- public ---

func (p *Parser) UpdateSources(ctx context.Context) error {
	slog.Debug("updating sources...")
	sources, err := p.parseAllSources(ctx)
	if err != nil {
		return err
	}
	p.sources = sources

	sources_all, err := database.GetSourcesExtended(ctx)
	if err != nil {
		slog.Warn("failed to update source statuses", "cause", err)
		return err
	}
	for source_type, sources := range sources_all {
		for source := range sources {
			if _, ok := p.sources[source_type][source]; !ok {
				// slog.Debug(fmt.Sprintf("set source '%s' to inactive", source))
				database.SetSourceStatus(ctx, source_type, source, "inactive")
			} else {
				database.SetSourceStatus(ctx, source_type, source, "active")
			}
		}
	}
	slog.Debug("done updating sources")
	return nil
}
func (p *Parser) GetSources() map[string]map[string]string {
	return p.sources
}
