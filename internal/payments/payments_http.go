package payments

import (
	"encoding/json"
	"log"
	"time"

	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/background"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/config"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/database"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/types"
	"github.com/araddon/dateparse"
	"github.com/gofiber/fiber/v2"
)

func Router(app *fiber.App) {

	app.Get("/", GetHealthChecks)

	app.Post("/payments", PostPayment)

	app.Delete("/cache", DeleteRedisCache)

	app.Get("/payments-summary", GetPaymentsSummary)
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

func DeleteRedisCache(c *fiber.Ctx) (err error) {
	err = database.RedisClient.Del(database.RedisCtx, config.RedisPaymentsKey).Err()
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return
	}

	c.Status(fiber.StatusOK)
	return nil
}

func GetPaymentsSummary(c *fiber.Ctx) (err error) {

	fromStr := c.Query("from")
	toStr := c.Query("to")
	from, to, err := parseTimeRange(fromStr, toStr)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return
	}
	paymentsData, err := database.RedisClient.HGetAll(database.RedisCtx, config.RedisPaymentsKey).Result()
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return
	}

	resp := summarizePayments(paymentsData, from, to)

	return c.Status(fiber.StatusOK).JSON(resp)
}

func parseTimeRange(fromStr string, toStr string) (from, to time.Time, err error) {

	if fromStr == "" && toStr == "" {
		return
	}
	if fromStr == "" || toStr == "" {
		return
	}
	from, err = dateparse.ParseAny(fromStr)
	if err != nil {
		return
	}
	to, err = dateparse.ParseAny(toStr)
	if err != nil {
		return
	}
	return
}

func summarizePayments(paymentsData map[string]string, from, to time.Time) types.PaymentsSummaryResponse {
	var defaultCount, fallbackCount int
	var defaultSum, fallbackSum float64

	isTimeRangeSet := !from.IsZero() && !to.IsZero()

	for _, paymentDataJson := range paymentsData {
		var payment types.ProcessedPayment
		if err := json.Unmarshal([]byte(paymentDataJson), &payment); err != nil {
			continue // Skip if JSON is malformed
		}

		requestedAt, err := time.Parse(time.RFC3339Nano, payment.RequestedAt)
		if err != nil {
			continue // Skip if date is malformed
		}

		if isTimeRangeSet && (requestedAt.Before(from) || requestedAt.After(to)) {
			continue
		}

		switch payment.PaymentProcessor {
		case config.PaymentProcessorUrlDefault:
			defaultCount++
			defaultSum += payment.Amount
		case config.PaymentProcessorUrlFallback:
			fallbackCount++
			fallbackSum += payment.Amount
		}
	}

	return types.PaymentsSummaryResponse{
		Default: types.PaymentsSummary{
			TotalRequests: defaultCount,
			TotalAmount:   defaultSum,
		},
		Fallback: types.PaymentsSummary{
			TotalRequests: fallbackCount,
			TotalAmount:   fallbackSum,
		},
	}
}
