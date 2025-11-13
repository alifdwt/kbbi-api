package scraper_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alifdwt/kbbi-api/internal/scraper"
)

func TestParseBahagiaFixture(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "tests", "fixtures", "bahagia.html")
	b, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	res, err := scraper.ParseFromHTML(string(b))
	if err != nil {
		t.Fatalf("ParseFromHTML error: %v", err)
	}

	// basic assertions
	if res.Kata != "bahagia" {
		t.Fatalf("expected kata=bahagia, got=%q", res.Kata)
	}
	if res.Pelafalan.Text != "ba·ha·gia" {
		t.Fatalf("expected pelafalan ba·ha·gia, got=%q", res.Pelafalan.Text)
	}
	if len(res.Entries) != 2 {
		t.Fatalf("expected 2 entries, got=%d", len(res.Entries))
	}

	// derived should contain three common lemmas
	wantDerived := map[string]bool{"ber·ba·ha·gia": false, "mem·ba·ha·gi·a·kan": false, "ke·ba·ha·gi·a·an": false}
	for _, d := range res.Derived {
		if _, ok := wantDerived[d.Kata]; ok {
			wantDerived[d.Kata] = true
			// senses should be non-empty and have teks
			if len(d.Senses) == 0 {
				t.Fatalf("derived %s has no senses", d.Kata)
			}
			for _, s := range d.Senses {
				if s.Teks == "" {
					t.Fatalf("derived %s has empty sense tekst", d.Kata)
				}
			}
		}
	}
	for k, found := range wantDerived {
		if !found {
			t.Fatalf("derived lemma %s not found", k)
		}
	}
}
