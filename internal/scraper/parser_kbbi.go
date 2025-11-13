package scraper

import (
	"encoding/json"
	"errors"
	"html"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alifdwt/kbbi-api/internal/model"
)

// ParseFromHTML improved:
// - split dHtml by double <br> to separate main senses and derived parts
// - parse main senses (headers) from first chunk
// - parse derived forms from subsequent chunks and put into result.Derived
// - clean and split examples properly
func ParseFromHTML(htmlPage string) (model.KBBIResult, error) {
	var result model.KBBIResult

	// extract jsdata textarea
	reTextarea := regexp.MustCompile(`(?s)<textarea[^>]*id=["']jsdata["'][^>]*>(.*?)</textarea>`)
	m := reTextarea.FindStringSubmatch(htmlPage)
	if len(m) < 2 {
		return result, errors.New("jsdata textarea not found")
	}
	raw := html.UnescapeString(m[1])

	// parse jsdata JSON array
	var arr []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return result, err
	}

	// collect memuat (x==5)
	for _, it := range arr {
		if xi, ok := it["x"]; ok {
			if asFloat, okf := xi.(float64); okf && int(asFloat) == 5 {
				if w, wok := it["w"].(string); wok {
					w = stripSup(w)
					result.Memuat = append(result.Memuat, map[string]string{
						"kata": w, "note": "tercantum sebagai turunan / memuat pada halaman",
					})
				}
			}
		}
	}

	// pick primary
	var primary map[string]interface{}
	for _, it := range arr {
		if xi, ok := it["x"]; ok {
			if asFloat, okf := xi.(float64); okf && int(asFloat) == 1 {
				primary = it
				break
			}
		}
	}
	if primary == nil && len(arr) > 0 {
		primary = arr[0]
	}
	if primary == nil {
		return result, errors.New("no primary entry found")
	}

	// basic fields
	if w, ok := primary["w"].(string); ok {
		result.Kata = stripSup(w)
	}
	result.Source = map[string]string{"site": "kbbi.web.id", "url": "https://kbbi.web.id/" + result.Kata}
	result.FetchedAt = time.Now()

	// extract pos mapping
	posMap := extractPOSMapping(htmlPage)

	// get dHtml
	var dHtml string
	if d, ok := primary["d"].(string); ok {
		dHtml = d
	}

	// pelafalan: first plausible <b>..</b>
	bRe := regexp.MustCompile(`(?i)<b>(.*?)</b>`)
	bMatches := bRe.FindAllStringSubmatch(dHtml, -1)
	if len(bMatches) > 0 {
		for _, bm := range bMatches {
			if len(bm) < 2 {
				continue
			}
			text := stripHTML(bm[1])
			if regexp.MustCompile(`^\d+$`).MatchString(text) {
				continue
			}
			if strings.Contains(text, "·") || (len(text) > 0 && len(text) < 60) {
				result.Pelafalan.Text = text
				result.Pelafalan.Raw = "/" + text + "/"
				break
			}
		}
	}

	// split parts by double <br> (assume first part: main senses; rest: derived)
	parts := regexp.MustCompile(`(?i)<br\s*/?>\s*<br\s*/?>`).Split(dHtml, -1)

	// parse main senses from parts[0]
	if len(parts) > 0 {
		main := parts[0]
		parseMainSenses(main, posMap, &result)
	}

	// parse derived forms from remaining parts
	if len(parts) > 1 {
		for _, p := range parts[1:] {
			parseDerivedPart(p, posMap, &result)
		}
	}

	// finalize notes
	if result.Notes == nil {
		result.Notes = map[string]interface{}{}
	}
	result.Notes["copyright"] = "Database utama menggunakan KBBI Daring edisi III; sumber kbbi.web.id"
	if _, ok := result.Notes["confidence"]; !ok {
		result.Notes["confidence"] = "high"
	}

	return result, nil
}

// parseMainSenses fills result.Entries using header positions
func parseMainSenses(dHtml string, posMap map[string]string, result *model.KBBIResult) {
	// headerRe finds <b>NUM</b> <em>CLS</em>
	headerRe := regexp.MustCompile(`(?i)<b>\s*(\d+)\s*</b>\s*<em>\s*([^<]+)\s*</em>\s*`)
	headers := headerRe.FindAllStringSubmatchIndex(dHtml, -1)
	if len(headers) == 0 {
		// fallback: whole chunk as one entry
		full := normalizeSpaces(stripHTML(dHtml))
		if full != "" {
			result.Entries = append(result.Entries, model.Entry{
				Nomor:     1,
				KelasKata: model.POS{Code: "", Label: ""},
				Teks:      full,
			})
		}
		return
	}

	for i, h := range headers {
		if len(h) < 6 {
			continue
		}
		numStr := dHtml[h[2]:h[3]]
		clsRaw := dHtml[h[4]:h[5]]
		num, _ := strconv.Atoi(strings.TrimSpace(numStr))
		cls := strings.TrimSpace(stripHTML(clsRaw))
		start := h[1]
		end := len(dHtml)
		if i+1 < len(headers) {
			end = headers[i+1][0]
		}
		chunk := dHtml[start:end]
		text := normalizeSpaces(stripHTML(chunk))
		examples := extractAndCleanExamples(chunk, cls)
		ent := model.Entry{
			Nomor:     num,
			KelasKata: model.POS{Code: cls, Label: posMap[cls]},
			Teks:      text,
			Contoh:    examples,
		}
		result.Entries = append(result.Entries, ent)
	}
}

// parseDerivedPart finds <b>word</b> occurrences and builds Derived entries
func parseDerivedPart(part string, posMap map[string]string, result *model.KBBIResult) {
	bRe := regexp.MustCompile(`(?i)<b>(.*?)</b>`)
	idxs := bRe.FindAllStringSubmatchIndex(part, -1)
	if len(idxs) == 0 {
		return
	}

	for i := 0; i < len(idxs); i++ {
		// capture lemma candidate
		cStart := idxs[i][2]
		cEnd := idxs[i][3]
		lemmaRaw := part[cStart:cEnd]
		lemma := stripHTML(lemmaRaw)

		// if this b is numeric, find next non-numeric b to treat as lemma and advance i
		if regexp.MustCompile(`^\d+$`).MatchString(strings.TrimSpace(lemma)) {
			found := false
			for j := i + 1; j < len(idxs); j++ {
				cs := idxs[j][2]
				ce := idxs[j][3]
				cand := stripHTML(part[cs:ce])
				if !regexp.MustCompile(`^\d+$`).MatchString(strings.TrimSpace(cand)) {
					lemma = cand
					i = j
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// determine chunk: from end of this lemma <b> to the start of next NON-NUMERIC <b> (next lemma), or end
		fullEnd := idxs[i][1]
		chunkStart := fullEnd
		chunkEnd := len(part)
		// find next non-numeric <b> index
		for k := i + 1; k < len(idxs); k++ {
			cs := idxs[k][2]
			ce := idxs[k][3]
			cand := stripHTML(part[cs:ce])
			if !regexp.MustCompile(`^\d+$`).MatchString(strings.TrimSpace(cand)) {
				// next lemma found; chunkEnds just before that <b>
				chunkEnd = idxs[k][0]
				break
			}
		}
		chunk := part[chunkStart:chunkEnd]

		// prefer <em> immediately after lemma as class marker
		leadingEmRe := regexp.MustCompile(`(?is)^\s*<em>\s*([^<]+)\s*</em>\s*`)
		cls := ""
		if lm := leadingEmRe.FindStringSubmatch(chunk); len(lm) > 1 {
			cls = strings.TrimSpace(stripHTML(lm[1]))
			// remove the leading em so it doesn't pollute teks
			chunk = leadingEmRe.ReplaceAllString(chunk, "")
		} else {
			// fallback: first short <em> or mapped one (remove only the first found)
			emRe := regexp.MustCompile(`(?i)<em>\s*([^<]+)\s*</em>`)
			if emIdx := emRe.FindStringSubmatchIndex(chunk); len(emIdx) > 1 {
				possible := strings.TrimSpace(stripHTML(chunk[emIdx[2]:emIdx[3]]))
				if len(possible) <= 5 || posMap[possible] != "" {
					cls = possible
					// remove only that first em occurrence by slicing
					chunk = chunk[:emIdx[0]] + chunk[emIdx[1]:]
				}
			}
		}

		// now within chunk, find numeric senses <b>1</b>, <b>2</b>, etc.
		numRe := regexp.MustCompile(`(?i)<b>\s*(\d+)\s*</b>\s*`)
		numIdxs := numRe.FindAllStringSubmatchIndex(chunk, -1)

		var senses []model.Entry
		if len(numIdxs) > 0 {
			for j, ni := range numIdxs {
				if len(ni) < 4 {
					continue
				}
				numStr := chunk[ni[2]:ni[3]]
				num, _ := strconv.Atoi(strings.TrimSpace(numStr))
				sStart := ni[1]
				sEnd := len(chunk)
				if j+1 < len(numIdxs) {
					sEnd = numIdxs[j+1][0]
				}
				sChunk := chunk[sStart:sEnd]

				text := normalizeSpaces(stripHTML(sChunk))
				text = stripLeadingClassToken(text)
				if text == "" {
					// fallback: loose text
					fallback := normalizeSpaces(regexp.MustCompile(`(?s)<[^>]*>`).ReplaceAllString(sChunk, " "))
					fallback = stripLeadingClassToken(fallback)
					text = strings.TrimSpace(fallback)
				}

				examples := extractAndCleanExamples(sChunk, cls)
				senses = append(senses, model.Entry{
					Nomor:     num,
					KelasKata: model.POS{Code: cls, Label: posMap[cls]},
					Teks:      text,
					Contoh:    examples,
				})
			}
		} else {
			// if no numeric senses inside chunk, use the whole chunk as a single sense
			text := normalizeSpaces(stripHTML(chunk))
			text = stripLeadingClassToken(text)
			if text == "" {
				text = normalizeSpaces(regexp.MustCompile(`(?s)<[^>]*>`).ReplaceAllString(chunk, " "))
				text = strings.TrimSpace(stripLeadingClassToken(text))
			}
			examples := extractAndCleanExamples(chunk, cls)
			senses = append(senses, model.Entry{
				Nomor:     1,
				KelasKata: model.POS{Code: cls, Label: posMap[cls]},
				Teks:      text,
				Contoh:    examples,
			})
		}

		// aggressive fallback: if all senses empty, take visible text from chunk (and trim lemma)
		allEmpty := true
		for _, s := range senses {
			if strings.TrimSpace(s.Teks) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			log.Printf("[parser] derived-empty -> lemma=%q chunk_raw=%q\n", lemma, chunk)
			fb := normalizeSpaces(regexp.MustCompile(`(?s)<[^>]*>`).ReplaceAllString(chunk, " "))
			fb = strings.TrimSpace(stripLeadingClassToken(fb))
			// remove lemma occurrences at beginning
			lowfb := strings.ToLower(fb)
			lowl := strings.ToLower(lemma)
			if strings.HasPrefix(lowfb, lowl) {
				fb = strings.TrimSpace(fb[len(lemma):])
			}
			if fb != "" {
				for si := range senses {
					if strings.TrimSpace(senses[si].Teks) == "" {
						senses[si].Teks = fb
					}
				}
			}
		}

		trimLemma := strings.TrimSpace(lemma)
		if trimLemma == "" || regexp.MustCompile(`^\d+$`).MatchString(trimLemma) {
			continue
		}
		d := model.Derived{
			Kata:      trimLemma,
			KelasKata: model.POS{Code: cls, Label: posMap[cls]},
			Senses:    senses,
		}
		result.Derived = append(result.Derived, d)
	}
}

// stripLeadingClassToken removes stray leading tokens like "a", "v", "n", "adv", etc.
func stripLeadingClassToken(s string) string {
	if s == "" {
		return s
	}
	return regexp.MustCompile(`(?i)^\s*(?:a|v|n|adv|pron|p|num|adj|v\.)\b[:\)\.\-\,\s]*`).ReplaceAllString(s, "")
}

// extract examples from a chunk, split by semicolon, filter invalid tokens and duplicates
func extractAndCleanExamples(chunk string, cls string) []string {
	exRe := regexp.MustCompile(`(?s)<em>(.*?)</em>`)
	matches := exRe.FindAllStringSubmatch(chunk, -1)
	out := []string{}
	seen := map[string]bool{}
	for _, mm := range matches {
		if len(mm) < 2 {
			continue
		}
		raw := stripHTML(mm[1])
		// split by semicolon and also by ' ; ' variations
		parts := regexp.MustCompile(`\s*;\s*`).Split(raw, -1)
		for _, p := range parts {
			item := strings.TrimSpace(p)
			// skip empty, skip item equal to cls (like 'n','a','v'), skip single-letter markers
			if item == "" {
				continue
			}
			low := strings.ToLower(item)
			if low == strings.ToLower(cls) {
				continue
			}
			// remove trailing punctuation
			item = strings.TrimRight(item, " .,:;")
			if len([]rune(item)) <= 1 {
				continue
			}
			// avoid items that are only non-alphanum
			if regexp.MustCompile(`^[^A-Za-z0-9]+$`).MatchString(item) {
				continue
			}
			if !seen[item] {
				out = append(out, item)
				seen[item] = true
			}
		}
	}
	return out
}

// helper utilities (same as before)

func stripSup(s string) string {
	re := regexp.MustCompile(`(?i)<sup.*?>.*?</sup>`)
	out := re.ReplaceAllString(s, "")
	out = stripHTML(out)
	return strings.TrimSpace(out)
}

func stripHTML(s string) string {
	re := regexp.MustCompile(`(?s)<[^>]*>`)
	out := re.ReplaceAllString(s, "")
	out = strings.ReplaceAll(out, "\u00b7", "·")
	out = strings.ReplaceAll(out, "&nbsp;", " ")
	out = strings.ReplaceAll(out, "&#183;", "·")
	out = strings.TrimSpace(out)
	return html.UnescapeString(out)
}

func normalizeSpaces(s string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(s), " ")
}

// extractPOSMapping unchanged from previous working version
func extractPOSMapping(htmlPage string) map[string]string {
	out := map[string]string{}

	idx := strings.Index(htmlPage, "_._4=")
	if idx == -1 {
		idx = strings.Index(htmlPage, "_._4 =")
		if idx == -1 {
			return out
		}
	}

	brStart := strings.Index(htmlPage[idx:], "{")
	if brStart == -1 {
		return out
	}
	brStart = idx + brStart
	level := 0
	endIdx := -1
	for i := brStart; i < len(htmlPage); i++ {
		switch htmlPage[i] {
		case '{':
			level++
		case '}':
			level--
			if level == 0 {
				endIdx = i
				break
			}
		}
	}
	if endIdx == -1 || endIdx <= brStart {
		return out
	}
	objText := htmlPage[brStart+1 : endIdx]

	pairRe := regexp.MustCompile(`([A-Za-z0-9_]+)\s*:\s*"(.*?)"\s*(,|$)`)
	matches := pairRe.FindAllStringSubmatch(objText, -1)
	for _, mm := range matches {
		if len(mm) >= 3 {
			k := strings.TrimSpace(mm[1])
			v := html.UnescapeString(strings.TrimSpace(mm[2]))
			v = regexp.MustCompile(`\s*\+.*`).ReplaceAllString(v, "")
			out[k] = v
		}
	}
	return out
}
