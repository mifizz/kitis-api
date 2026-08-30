package parser

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"strings"

	api_errors "kitis-api/internal/errors"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

func (p *Parser) fetchDoc(ctx context.Context, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err == context.DeadlineExceeded {
		return nil, api_errors.ErrHTTPTimedOut
	} else if err != nil {
		return nil, err
	} else if resp.StatusCode == 404 {
		return nil, api_errors.ErrParserPageNotFound
	}

	defer resp.Body.Close()
	utf8_body, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(utf8_body)
}

func fixNewlines(doc *goquery.Document) {
	doc.Find("br").Each(func(i int, s *goquery.Selection) {
		s.ReplaceWithHtml("\\n")
	})
}

func parseModifiedTime(doc *goquery.Document) int64 {
	return parseModifiedTimeString(strings.TrimSpace(doc.Find("div.ref").Text()))
}

func selToInt(s *goquery.Selection) int {
	text := s.Text()
	number, _ := strconv.Atoi(text)
	return number
}
func selToFloat(s *goquery.Selection, precision int) float64 {
	text := strings.Replace(s.Text(), ",", ".", 1)
	number, _ := strconv.ParseFloat(text, 64)
	number = math.Round(number*math.Pow10(precision)) / float64(math.Pow10(precision))
	return number
}
