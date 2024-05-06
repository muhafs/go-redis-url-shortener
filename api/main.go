package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/muhafs/go-redis-url-shortener/routes"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println(err)
	}

	app := fiber.New()
	app.Use(logger.New())

	setupRoutes(app)

	if err := app.Listen(os.Getenv("APP_PORT")); err != nil {
		log.Fatal(err)
	}
}
