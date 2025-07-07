package config

import "os"

var DatabaseURL string
var CacheURL string
var PaymentProcessorUrlDefault string
var PaymentProcessorUrlFallback string

func init() {
	DatabaseURL = os.Getenv("DB_CONNECTION_STRING")
	CacheURL = os.Getenv("REDIS_ADDRESS")
	PaymentProcessorUrlDefault = os.Getenv("PAYMENT_PROCESSOR_URL_DEFAULT")
	PaymentProcessorUrlFallback = os.Getenv("PAYMENT_PROCESSOR_URL_FALLBACK")
}
