package main

import (
	"log"

	paymentHttp "github.com/JoaoReisA/rinha-de-backend-2025-go/internal/paymentsHttp"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	paymentHttp.Handler(app)

	log.Fatal(app.Listen(":8080"))
}
