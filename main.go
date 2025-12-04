package main

import (
	"log"

	"github.com/alifdwt/kbbi-api/config"
	"github.com/alifdwt/kbbi-api/database"
	"github.com/alifdwt/kbbi-api/handlers"
	"github.com/alifdwt/kbbi-api/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Initialize database
	dbConfig := config.NewDatabaseConfig()
	db, err := config.ConnectDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	dictDB := database.NewDictionaryDB(db)
	if err := dictDB.TestConnection(); err != nil {
		log.Fatalf("Database connection test failed: %v", err)
	}

	// Initialize handlers
	dictHandler := handlers.NewDictionaryHandler(dictDB)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "KBBI API",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Setup routes
	routes.SetupRoutes(app, dictHandler)

	// Get port from environment or use default 8080
	port := config.GetEnv("PORT", "8080")

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}