package handlers

import (
	"math"
	"strconv"

	"github.com/alifdwt/kbbi-api/models"
	"github.com/gofiber/fiber/v2"
)

type DictionaryHandler struct {
	dictDB DictionaryDBInterface
}

type DictionaryDBInterface interface {
	SearchWords(query string, limit, offset int) ([]models.Dictionary, error)
	GetWordByExactMatch(word string) (*models.Dictionary, error)
	CountWords(query string) (int64, error)
	GetTotalWordsCount() (int64, error)
	TestConnection() error
}

func NewDictionaryHandler(dictDB DictionaryDBInterface) *DictionaryHandler {
	return &DictionaryHandler{dictDB: dictDB}
}

func (h *DictionaryHandler) SearchWords(c *fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Query parameter is required",
		})
	}

	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := (page - 1) * limit

	words, err := h.dictDB.SearchWords(query, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to search words",
		})
	}

	total, err := h.dictDB.CountWords(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to count results",
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response := models.SearchResponse{
		Data:       make([]models.DictionaryResponse, len(words)),
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	for i, word := range words {
		response.Data[i] = models.DictionaryResponse{
			Word: word.Word,
			Arti: word.Arti,
			Type: word.Type,
		}
	}

	return c.JSON(response)
}

func (h *DictionaryHandler) GetWord(c *fiber.Ctx) error {
	word := c.Params("word")
	if word == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Word parameter is required",
		})
	}

	result, err := h.dictDB.GetWordByExactMatch(word)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get word",
		})
	}

	if result == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Word not found",
		})
	}

	response := models.DictionaryResponse{
		Word: result.Word,
		Arti: result.Arti,
		Type: result.Type,
	}

	return c.JSON(response)
}

func (h *DictionaryHandler) GetStats(c *fiber.Ctx) error {
	total, err := h.dictDB.GetTotalWordsCount()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get stats",
		})
	}

	return c.JSON(fiber.Map{
		"total_words": total,
	})
}