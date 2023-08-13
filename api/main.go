package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/thepranays/url-shortner-gofibre-redis/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error Loading env file")
	}
	app := fiber.New()    //creating instance of app 'similar to express'
	app.Use(logger.New()) //attaching logger
	setupRoutes(app)
	log.Fatal(app.Listen(os.Getenv("API_PORT"))) //start server ,if fails shows log
}

// Setup all routes
func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)
}
