package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/alifdwt/kbbi-api/models"
	"github.com/jmoiron/sqlx"
)

type DictionaryDB struct {
	db *sqlx.DB
}

func NewDictionaryDB(db *sqlx.DB) *DictionaryDB {
	return &DictionaryDB{db: db}
}

func (d *DictionaryDB) SearchWords(query string, limit, offset int) ([]models.Dictionary, error) {
	var words []models.Dictionary

	err := d.db.Select(&words, `
		SELECT id, word, arti, type, created_at, updated_at
		FROM dictionary
		WHERE word ILIKE $1
		ORDER BY word ASC
		LIMIT $2 OFFSET $3`,
		fmt.Sprintf("%%%s%%", query), limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to search words: %w", err)
	}

	return words, nil
}

func (d *DictionaryDB) GetWordByExactMatch(word string) (*models.Dictionary, error) {
	var dict models.Dictionary

	err := d.db.Get(&dict, `
		SELECT id, word, arti, type, created_at, updated_at
		FROM dictionary
		WHERE LOWER(word) = LOWER($1)`,
		word)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get word: %w", err)
	}

	return &dict, nil
}

func (d *DictionaryDB) CountWords(query string) (int64, error) {
	var count int64

	err := d.db.Get(&count, `
		SELECT COUNT(*)
		FROM dictionary
		WHERE word ILIKE $1`,
		fmt.Sprintf("%%%s%%", query))

	if err != nil {
		return 0, fmt.Errorf("failed to count words: %w", err)
	}

	return count, nil
}

func (d *DictionaryDB) GetTotalWordsCount() (int64, error) {
	var count int64

	err := d.db.Get(&count, "SELECT COUNT(*) FROM dictionary")
	if err != nil {
		return 0, fmt.Errorf("failed to get total words count: %w", err)
	}

	return count, nil
}

func (d *DictionaryDB) TestConnection() error {
	var result int
	err := d.db.Get(&result, "SELECT 1")
	if err != nil {
		return fmt.Errorf("database connection test failed: %w", err)
	}

	log.Println("Database connection successful")
	return nil
}