package scraper

import (
	"encoding/json"
	"errors"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alifdwt/kbbi-api/internal/model"
)

func ParseFromHTML(htmlPage string) (model.KBBIResult, error) {
	var result model.KBBIResult

	re := regexp.MustCompile(`(?s)<textarea[^>]*id="jsdata"[^>]*>(.*?)</textarea>`)
	m := re.FindStringSubmatch(htmlPage)
	if len(m) < 2 {
		return result, errors.New("jsdata textarea not found")
	}

	raw := m[1]
	raw = html.UnescapeString(raw)

	var arr []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return result, err
	}

	var primary map[string]interface{}
	for _, item := range arr {
		if xi, ok := item["x"]; ok {
			switch v := xi.(type) {
			case float64:
				if int(v) == 1 {
					primary = item
					break
				}
			case int:
				if v == 1 {
					primary = item
					break
				}
			}
		}
	}

	if primary == nil && len(arr) > 0 {
		primary = arr[0]
	}
	if primary == nil {
		return result, errors.New("no entries in jsdata")
	}

	if w, ok := primary["w"].(string); ok {
		w = regexp.MustCompile(`<sup>.*?</sup>`).ReplaceAllString(w, "")
		w = strings.TrimSpace(w)
		result.Kata = w
	}
	result.Source = map[string]string{"site": "kbbi.web.id", "url": "https://kbbi.web.id/" + result.Kata}
	result.FetchedAt = time.Now()

	var dHtml string
	if d, ok := primary["d"].(string); ok {
		dHtml = d
	}

	parts := regexp.MustCompile(`(?i)<br\s*/?>\s*<br\s*/?>`).Split(dHtml, -1)
	if len(parts) > 0 {
		// entryRe := regexp.MustCompile(`(?s)<b>\s*(?:[0-9]+)\s*</b>\s*<em>(.*?)</em>\s*(.*?)$`)

		numRe := regexp.MustCompile(`(?s)<b>(\d+)</b>\s*<em>([^<]+)</em>\s*([^<].*?)(?:$|<br\s*/?>)`)
		nums := numRe.FindAllStringSubmatch(parts[0], -1)
		for _, g := range nums {
			if len(g) >= 4 {
				numStr := g[1]
				cls := strings.TrimSpace(g[2])
				text := strings.TrimSpace(stripHTML(g[3]))
				num, _ := strconv.Atoi(numStr)
				e := model.Entry{
					Nomor: num,
					KelasKata: model.POS{
						Code:  cls,
						Label: posLabel(cls),
					},
					Teks: text,
				}

				exampleRe := regexp.MustCompile(`<em>(.*?)</em>`)
				exs := exampleRe.FindAllStringSubmatch(parts[0], -1)
				if len(exs) > 0 {
					for _, ex := range exs {
						if len(ex) > 1 {
							e.Contoh = append(e.Contoh, strings.TrimSpace(stripHTML(ex[1])))
						}
					}
				}
				result.Entries = append(result.Entries, e)
			}
		}

		if len(result.Entries) == 0 {
			text := stripHTML(parts[0])
			e := model.Entry{
				Nomor: 1,
				KelasKata: model.POS{
					Code:  "",
					Label: "",
				},
				Teks: strings.TrimSpace(text),
			}
			result.Entries = append(result.Entries, e)
		}
	}

	if len(parts) > 1 {
		bRe := regexp.MustCompile(`<b>([^<]+)</b>\s*<em>([^<]+)</em>\s*<b>(\d+)</b>\s*(.*?)($|<br\s*/?>)`)
		derMatches := bRe.FindAllStringSubmatch(strings.Join(parts[1:], "\n"), -1)
		for _, dm := range derMatches {
			if len(dm) >= 5 {
				kata := stripHTML(dm[1])
				cls := strings.TrimSpace(dm[2])
				d := model.Derived{
					Kata: kata,
					KelasKata: model.POS{
						Code:  cls,
						Label: posLabel(cls),
					},
				}
				result.Derived = append(result.Derived, d)
			}
		}
	}

	if result.Kata == "" {
		return result, errors.New("failed to determine kata")
	}

	result.Notes = map[string]any{"copyright": "Database utama menggunakan KBBI Daring edisi III; sumber kbbi.web.id", "confidence": "medium"}

	return result, nil
}

func stripHTML(s string) string {
	re := regexp.MustCompile(`(?s)<[^>]*>`)
	out := re.ReplaceAllString(s, "")
	out = strings.ReplaceAll(out, "&nbsp;", " ")
	out = strings.TrimSpace(out)
	return out
}

func posLabel(code string) string {
	m := map[string]string{
		"n": "nomina (kata benda)",
		"a": "adjektiva (kata yang menjelaskan nomina atau Pronomina)",
		"v": "verba (kata kerja)",
	}
	if v, ok := m[code]; ok {
		return v
	}
	return ""
}
