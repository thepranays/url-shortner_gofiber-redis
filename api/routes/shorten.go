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
	"github.com/thepranays/url-shortner-gofibre-redis/database"
	"github.com/thepranays/url-shortner-gofibre-redis/helpers"
)

type request struct {
	URL         string        `json:"url"`   //as go doesnt understand json so
	CustomShort string        `json:"short"` // it has do alot encoding decoding(serialization),so we have to mention how it looks like
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`       //Limit for service request
	XRateLimitReset time.Duration `json:"rate_limit_reset"` //Reset time period for the limit

}

func ShortenURL(ctx *fiber.Ctx) error {
	//Request
	body := request{} // or new(request)
	fmt.Println("hi")
	if err := ctx.BodyParser(&body); err != nil {
		ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Cannot parse json into struct"})
		return fiber.NewError(401, "Invalid Body")
	} //to parse json to struct

	//Rate limiter
	rdb_rate := database.CreateClient(1)
	defer rdb_rate.Close()
	value, err := rdb_rate.Get(database.Ctx, ctx.IP()).Result()
	if err == redis.Nil {
		_ = rdb_rate.Set(database.Ctx, ctx.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err() //returns error if occured

	} else {
		valueInt, _ := strconv.Atoi(value)
		resetTimeLeft, _ := rdb_rate.TTL(database.Ctx, ctx.IP()).Result()
		if valueInt <= 0 {
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"Error": "Service Rate Limit exceed", "reset_time_left": resetTimeLeft / time.Second / time.Minute})
		}

	}

	//CHECK:URL Validator
	if !govalidator.IsURL(body.URL) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid URL"})
	}

	//CHECK:Domain Error to ensure loop doesnt occur if someone shorts the endpoint itself
	if !helpers.RemoveDomainError(body.URL) {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"Error": "Unavailable Service ;_;"})
	}
	//enforce http
	body.URL = helpers.EnforceHTTP(body.URL)

	//Custom URL
	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}
	rdb_url := database.CreateClient(0)
	defer rdb_url.Close()
	_, err = rdb_url.Get(database.Ctx, id).Result()
	fmt.Println(err)
	if err != redis.Nil {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Custom Short already exists"})
	}
	if body.Expiry == 0 { //if url expiry not provided DEFAULT:24hr
		body.Expiry = 24
	}
	err = rdb_url.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err() //storing short in db
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "FAILED:Could not store short url"})
	}

	rdb_rate.Decr(database.Ctx, ctx.IP()) //decrement API quota count of current ip

	//Response
	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	val, _ := rdb_rate.Get(database.Ctx, ctx.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)
	ttl, _ := rdb_rate.TTL(database.Ctx, ctx.IP()).Result()
	resp.XRateLimitReset = (ttl / time.Nanosecond / time.Minute)
	resp.CustomShort = os.Getenv("API_DOMAIN") + "/" + id //shorten url

	return ctx.Status(fiber.StatusOK).JSON(resp) //return response
}
