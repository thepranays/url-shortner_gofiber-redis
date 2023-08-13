package routes

import (
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
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
	body := request{} // or new(request)

	if err := ctx.BodyParser(&body); err != nil {
		ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Cannot parse json into struct"})
		return fiber.NewError(401, "Invalid Body")
	} //to parse json to struct

	//Middleware:Rate limiter

	//CHECK:URL Validator
	if !govalidator.IsURL(body.URL) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid URL"})
	}

	//CHECK:Domain Error
	if !helpers.RemoveDomainError(body.URL) {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"Error": "Unavailable Service ;_;"})
	}
	//enforce https (SSL)
	body.URL = helpers.EnforceHTTPS(body.URL)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"Success": "200"})
}
