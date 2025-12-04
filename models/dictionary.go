package models

import "time"

type Dictionary struct {
	ID        int       `json:"id" db:"id"`
	Word      string    `json:"word" db:"word"`
	Arti      string    `json:"arti" db:"arti"`
	Type      int       `json:"type" db:"type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type DictionaryResponse struct {
	Word string `json:"word"`
	Arti string `json:"arti"`
	Type int    `json:"type"`
}

type SearchRequest struct {
	Query string `json:"query" query:"query"`
	Page  int    `json:"page" query:"page"`
	Limit int    `json:"limit" query:"limit"`
}

type SearchResponse struct {
	Data       []DictionaryResponse `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"total_pages"`
}