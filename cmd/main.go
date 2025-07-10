package main

import (
	"log"

	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/background"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/database"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/payments"
	"github.com/gofiber/fiber/v2"
)

func main() {
	database.ConnectRedis()

	app := fiber.New()
	payments.Router(app)

	go background.RunHealthCheckWorker()
	background.StartWorker()

	log.Fatal(app.Listen(":8080"))
}
