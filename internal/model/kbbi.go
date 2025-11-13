package model

import "time"

type POS struct {
	Code  string `json:"code"`
	Label string `json:"label,omitempty"`
}

type Entry struct {
	Nomor     int      `json:"nomor"`
	KelasKata POS      `json:"kelas_kata"`
	Teks      string   `json:"teks"`
	Contoh    []string `json:"contoh,omitempty"`
}

type Derived struct {
	Kata      string  `json:"kata"`
	KelasKata POS     `json:"kelas_kata"`
	Senses    []Entry `json:"senses,omitempty"`
}

type KBBIResult struct {
	Kata      string `json:"kata"`
	Pelafalan struct {
		Text string `json:"text,omitempty"`
		Raw  string `json:"raw,omitempty"`
	} `json:"pelafalan"`
	Source    map[string]string   `json:"source,omitempty"`
	FetchedAt time.Time           `json:"fetched_at"`
	Entries   []Entry             `json:"entries,omitempty"`
	Derived   []Derived           `json:"derived,omitempty"`
	Memuat    []map[string]string `json:"memuat,omitempty"`
	Notes     map[string]any      `json:"notes,omitempty"`
}
