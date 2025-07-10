package background

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/config"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/database"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/types"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func ProccessPayments(paymentRequest types.PaymentRequest) (err error) {
	requestBody, err := json.Marshal(paymentRequest)
	if err != nil {
		return err
	}
	paymentProcessorUsed := HealthCache.BestPaymentProcessorUrl
	agent := fiber.Post(paymentProcessorUsed + "/payments")
	agent.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	agent.Body(requestBody)
	statusCode, _, errs := agent.Bytes()
	if errs != nil {
		log.Printf("Error calling payment processor: %v", errs)
		return errs[0]
	}
	if statusCode < 200 || statusCode >= 300 {
		return utils.ErrFailedToSendPaymentProcess
	}
	processedPayment := types.ProcessedPayment{
		CorrelationID:    paymentRequest.CorrelationId,
		Amount:           paymentRequest.Amount,
		PaymentProcessor: paymentProcessorUsed,
		RequestedAt:      paymentRequest.RequestedAt,
	}

	paymentData, err := json.Marshal(processedPayment)
	if err != nil {
		return fmt.Errorf("failed to marshal payment data: %w", err)
	}
	err = database.RedisClient.HSet(database.RedisCtx, config.RedisPaymentsKey, processedPayment.CorrelationID, paymentData).Err()
	if err != nil {
		return fmt.Errorf("failed to save payment in redis: %w", err)
	}
	log.Println("PAYMENT SAVED ON REDIS")
	//TODO: CALL WORKER TO INSERT In BATCH ON POSTGRES

	return nil
}
