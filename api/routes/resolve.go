package routes

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/thepranays/url-shortner-gofibre-redis/database"
)

func ResolveURL(ctx *fiber.Ctx) error {
	url := ctx.Params("url")
	rdb := database.CreateClient(0)
	defer rdb.Close()
	value, err := rdb.Get(database.Ctx, url).Result() //Redis is key-value pair database (used for caching)
	if err == redis.Nil {                             //not found in db
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"Error": "Url not found"})
	} else if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Cannot connect to database service"})

	}
	//Increment counter (no. of resolves)
	resolveInr := database.CreateClient(1)
	defer resolveInr.Close()
	_ = resolveInr.Incr(database.Ctx, "counter")

	return ctx.Redirect(value, 301)
}
