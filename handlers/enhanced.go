package handlers

import (
	"math"
	"strconv"

	"github.com/alifdwt/kbbi-api/models"
	"github.com/gofiber/fiber/v2"
)

func (h *DictionaryHandler) GetWordEnhanced(c *fiber.Ctx) error {
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

	// Parse the text into structured format
	enhanced := models.ParseDictionaryText(result.Arti)

	// Override the parsed word with the actual word from database
	enhanced.Word = result.Word

	return c.JSON(enhanced)
}

func (h *DictionaryHandler) SearchWordsEnhanced(c *fiber.Ctx) error {
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

	// Parse all words to enhanced format
	var enhancedResults []models.EnhancedDictionary
	for _, word := range words {
		enhanced := models.ParseDictionaryText(word.Arti)
		enhanced.Word = word.Word
		enhancedResults = append(enhancedResults, enhanced)
	}

	response := models.EnhancedSearchResponse{
		Data:       enhancedResults,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	return c.JSON(response)
}