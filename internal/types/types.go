package types

import (
	"math"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type HealthStatusCache struct {
	Mu                      sync.RWMutex
	Status                  fiber.Map
	BestPaymentProcessorUrl string
	Err                     error
}
type HealthResult struct {
	Response fiber.Map
	Err      error
}

type ProcessorHealth struct {
	IsFailing       bool    `json:"failing"`
	MinResponseTime float64 `json:"min_response_time"` // Included as requested
}
type HealthStatusPayload struct {
	Default  ProcessorHealth `json:"default"`
	Fallback ProcessorHealth `json:"fallback"`
}

type PaymentRequest struct {
	CorrelationId string  `json:"correlationId" xml:"correlationId" form:"correlationId"`
	Amount        float64 `json:"amount" xml:"amount" form:"amount"`
	RequestedAt   string  `json:"requestedAt" xml:"requestedAt" form:"requestedAt"`
}

type ProcessedPayment struct {
	CorrelationID    string  `json:"correlationId"`
	Amount           float64 `json:"amount"`
	PaymentProcessor string  `json:"paymentProcessor"`
	RequestedAt      string  `json:"requestedAt"`
}

type PaymentsSummary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentsSummaryResponse struct {
	Default  PaymentsSummary `json:"default"`
	Fallback PaymentsSummary `json:"fallback"`
}

func RoundFloat(val float64) float64 {
	return math.Round(float64(val)*10) / 10
}
