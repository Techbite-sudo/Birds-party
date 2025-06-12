package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/JILI-GAMES/b_backend_games8/pkg/common/config"
	"github.com/JILI-GAMES/b_backend_games8/pkg/common/rng"
	"github.com/JILI-GAMES/b_backend_games8/pkg/common/settings"
	"github.com/JILI-GAMES/b_backend_games8/pkg/games/birdsparty"
)

func main() {
	// Load configuration
	prodCfg, testCfg := config.LoadAll()

	fmt.Println("Production Configuration:", prodCfg)
	fmt.Println("Test Configuration:", testCfg)

	// Set up logging
	logFile, err := os.OpenFile(prodCfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Create shared clients
	rngClient := rng.NewClient(prodCfg.RNGServiceURL)
	settingsClient := settings.NewClient(prodCfg.SettingsServiceURL)
	
	// Create test clients
	rngTestClient := rng.NewClient(testCfg.RNGServiceURL)
	settingsTestClient := settings.NewClient(testCfg.SettingsServiceURL)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		Output:     logFile,
	}))

	// Register routes for Birds Party
	birdsPartyRoutes := birdsparty.NewRouteGroup(rngClient, settingsClient, rngTestClient, settingsTestClient)
	birdsPartyRoutes.Register(app)

	// Add a simple status endpoint
	app.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"game":   "birdsparty",
		})
	})

	// Start the server
	port := prodCfg.ServerPort
	log.Printf("Starting Birds Party server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

// Custom error handler
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	// Handle common errors with appropriate status codes
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	log.Printf("Request failed: %v", err)
	return c.Status(code).JSON(fiber.Map{
		"status":  "error",
		"message": err.Error(),
	})
}