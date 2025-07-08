package main

import (
	"log"

	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/background"
	paymentHttp "github.com/JoaoReisA/rinha-de-backend-2025-go/internal/paymentsHttp"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	go background.RunHealthCheckWorker()
	paymentHttp.Handler(app)

	log.Fatal(app.Listen(":8080"))
}
