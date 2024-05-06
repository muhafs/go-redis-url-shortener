package routes

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muhafs/go-redis-url-shortener/database"
	"github.com/muhafs/go-redis-url-shortener/helper"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "cannot parse JSON",
		})
	}

	//implement rate limiting
	r2 := database.CreateClient(1)
	defer r2.Close()

	_, err := r2.Get(database.CTX, c.IP()).Result()
	if err == redis.Nil {
		r2.Set(database.CTX, c.IP(), os.Getenv("API_QUOTA"), 30*time.Minute).Err()
	} else {
		val, _ := r2.Get(database.CTX, c.IP()).Result()
		valInt, _ := strconv.Atoi(val)

		if valInt <= 0 {
			limit, _ := r2.TTL(database.CTX, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"message":          "rate limit exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	//check if the input is an actual URL
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid URL",
		})
	}

	//check for domain error
	if !helper.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"message": "oops, invalid request!",
		})
	}

	//enforce https/SSL
	body.URL = helper.EnforceHTTP(body.URL)

	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()

	val, _ := r.Get(database.CTX, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "custom short already used",
		})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	if err = r.Set(database.CTX, id, body.URL, body.Expiry*time.Hour).Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "unable to connect to server",
		})
	}

	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  100,
		XRateLimitReset: 100,
	}

	r2.Decr(database.CTX, c.IP())
	val, _ = r2.Get(database.CTX, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(database.CTX, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = fmt.Sprintf("%v/%v", os.Getenv("DOMAIN"), id)

	return c.Status(fiber.StatusOK).JSON(resp)
}
