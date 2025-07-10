package payments

import (
	"encoding/json"
	"log"
	"time"

	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/background"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/config"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/database"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/types"
	"github.com/gofiber/fiber/v2"
)

func Router(app *fiber.App) {

	app.Get("/", GetHealthChecks)

	app.Post("/payments", PostPayment)
}

func GetHealthChecks(c *fiber.Ctx) (err error) {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":        background.HealthCache.Status,
		"bestProcessor": background.HealthCache.BestPaymentProcessorUrl,
	})
}

func PostPayment(c *fiber.Ctx) (err error) {
	paymentRequest := new(types.PaymentRequest)
	if err := c.BodyParser(paymentRequest); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	currentTime := time.Now().UTC()
	formattedString := currentTime.Format(time.RFC3339Nano)
	paymentRequest.RequestedAt = formattedString
	jsonBody, err := json.Marshal(paymentRequest)
	if err != nil {
		log.Printf("ERROR ON MARSHAL TO JSON: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if err := database.RedisClient.LPush(database.RedisCtx, config.RedisQueueKey, jsonBody).Err(); err != nil {
		log.Printf("ERROR ON PUSH TO REDIS QUEUE: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusCreated)
}
