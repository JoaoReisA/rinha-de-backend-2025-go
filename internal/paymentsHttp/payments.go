package paymentsHttp

import (
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/config"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/client"
)

func Handler(app *fiber.App) {

	app.Get("/", func(c fiber.Ctx) error {
		cc := client.New()

		resp, err := cc.Get(config.PaymentProcessorUrlDefault + "/payments/service-health")
		if err != nil {
			panic(err)
		}
		return c.SendString(resp.String())
	})

	app.Post("/payments", func(c fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})
}

func PostPayment(c *fiber.Ctx) {
	cc := client.New()

	resp, err := cc.Post(config.PaymentProcessorUrlDefault, client.Config{})
	if err != nil {
		panic(err)
	}
	print(resp)
}
