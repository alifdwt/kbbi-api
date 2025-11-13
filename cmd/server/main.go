package main

import (
	"log"
	"os"

	"github.com/alifdwt/kbbi-api/internal/api"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()
	app.Use(logger.New())

	api.RegisterRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("starting server :%s...", port)
	log.Fatal(app.Listen(":" + port))
}
