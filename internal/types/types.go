package types

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
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
	CorrelationId string          `json:"correlationId" xml:"correlationId" form:"correlationId"`
	Amount        decimal.Decimal `json:"amount" xml:"amount" form:"amount"`
	RequestedAt   string          `json:"requestedAt" xml:"requestedAt" form:"requestedAt"`
}

type ProcessedPayment struct {
	CorrelationID    string          `json:"correlationId"`
	Amount           decimal.Decimal `json:"amount"`
	PaymentProcessor string          `json:"paymentProcessor"`
	RequestedAt      string          `json:"requestedAt"`
}
