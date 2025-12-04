package routes

import (
	"github.com/alifdwt/kbbi-api/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, dictHandler *handlers.DictionaryHandler) {
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
		})
	})

	// Dictionary routes
	api.Get("/search", dictHandler.SearchWords)
	api.Get("/word/:word", dictHandler.GetWord)
	api.Get("/stats", dictHandler.GetStats)

	// Root route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "KBBI API",
			"version": "1.0.0",
			"endpoints": fiber.Map{
				"search": "/api/v1/search?query=<word>&page=1&limit=20",
				"word":   "/api/v1/word/<word>",
				"stats":  "/api/v1/stats",
				"health": "/api/v1/health",
			},
		})
	})
}