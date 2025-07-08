package paymentsHttp

import (
	"encoding/json"

	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/background"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/config"
	"github.com/gofiber/fiber/v2"
)

func Handler(app *fiber.App) {

	app.Get("/", GetHealthChecks)

	app.Post("/payments", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})
}

func GetHealthChecks(c *fiber.Ctx) (err error) {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":        background.HealthCache.Status,
		"bestProcessor": background.HealthCache.BestPaymentProcessorUrl,
	})
}

func PostPayment(c *fiber.Ctx) (err error) {

	agent := fiber.Post(config.PaymentProcessorUrlDefault)
	statusCode, body, errs := agent.Bytes()
	if errs != nil {
		panic(errs)
	}
	var something fiber.Map
	err = json.Unmarshal(body, &something)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err,
		})
	}

	return c.Status(statusCode).JSON(something)

}
