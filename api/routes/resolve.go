package routes

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/muhafs/go-redis-url-shortener/database"
)

func ResolveURL(c *fiber.Ctx) error {
	url := c.Params("url")

	r := database.CreateClient(0)
	defer r.Close()

	res, err := r.Get(database.CTX, url).Result()

	if err != redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "short url not found",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "cannot connect to database",
		})
	}

	rInr := database.CreateClient(1)
	defer rInr.Close()

	_ = rInr.Incr(database.CTX, "counter")

	return c.Redirect(res, 301)
}
