package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/alifdwt/kbbi-api/internal/scraper"
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	app.Get("/api/kata/:lema", getKataHandler)
	app.Get("/health", func(c *fiber.Ctx) error { return c.SendString("ok") })
}

func getKataHandler(c *fiber.Ctx) error {
	lema := c.Params("lema")
	if lema == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "missing lema",
		})
	}

	if lema != "bahagia" {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "not found (fixture only)",
		})
	}

	path := filepath.Join("tests", "fixtures", "bahagia.html")
	b, err := os.ReadFile(path)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to read fixture",
		})
	}

	res, err := scraper.ParseFromHTML(string(b))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(res)
}
