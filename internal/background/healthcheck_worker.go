package background

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/config"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/types"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/utils"
	"github.com/gofiber/fiber/v2"
)

var HealthCache = types.HealthStatusCache{
	Status:                  make(fiber.Map),
	BestPaymentProcessorUrl: config.PaymentProcessorUrlDefault,
	Err:                     nil,
}

func RunHealthCheckWorker() {
	log.Println("Health check worker started...")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		log.Println("Worker: Running health checks...")

		newStatus, err := fetchAndProcessHealthChecks()
		if err != nil {
			log.Printf("Worker: Error fetching health checks: %v", err)
			HealthCache.Err = err
			continue // Skip this update and wait for the next tick
		}

		// Safely write the new status to the global cache
		HealthCache.Mu.Lock()
		HealthCache.Status = newStatus
		HealthCache.BestPaymentProcessorUrl, HealthCache.Err = decideBestProcessor(newStatus)
		HealthCache.Mu.Unlock()

		log.Println("Worker: Health status cache updated successfully.")
	}
}

func decideBestProcessor(m fiber.Map) (string, error) {
	var payload types.HealthStatusPayload

	jsonBytes, err := json.Marshal(m)
	if err != nil {
		log.Printf("Error converting health status map to JSON: %v", err)
		return config.PaymentProcessorUrlDefault, err
	}

	if err := json.Unmarshal(jsonBytes, &payload); err != nil {
		log.Printf("Error parsing health status structure: %v", err)
		return config.PaymentProcessorUrlDefault, err
	}

	if payload.Default.IsFailing && payload.Fallback.IsFailing {
		log.Println("CRITICAL: Both default and fallback payment processors are failing.")
		return config.PaymentProcessorUrlDefault, utils.ErrAllPaymentsFailing
	}

	if payload.Default.IsFailing || payload.Default.MinResponseTime > payload.Fallback.MinResponseTime+50 {
		return config.PaymentProcessorUrlFallback, nil
	}
	return config.PaymentProcessorUrlDefault, nil
}
func fetchAndProcessHealthChecks() (fiber.Map, error) {
	var wg sync.WaitGroup
	defaultChan := make(chan types.HealthResult, 1)
	fallbackChan := make(chan types.HealthResult, 1)

	wg.Add(2)
	go fetchHealth(config.PaymentProcessorUrlDefault+"/payments/service-health", &wg, defaultChan)
	go fetchHealth(config.PaymentProcessorUrlFallback+"/payments/service-health", &wg, fallbackChan)

	wg.Wait()
	close(defaultChan)
	close(fallbackChan)

	defaultResult := <-defaultChan
	fallbackResult := <-fallbackChan

	if defaultResult.Err != nil {
		return nil, defaultResult.Err
	}
	if fallbackResult.Err != nil {
		return nil, fallbackResult.Err
	}

	response := fiber.Map{
		"default":          defaultResult.Response,
		"fallback":         fallbackResult.Response,
		"last_checked_utc": time.Now().UTC(),
	}

	return response, nil
}
func fetchHealth(url string, wg *sync.WaitGroup, ch chan<- types.HealthResult) {
	defer wg.Done()
	agent := fiber.Get(url)
	_, body, errs := agent.Bytes()
	if len(errs) > 0 {
		ch <- types.HealthResult{Err: errs[0]}
		return
	}
	var res fiber.Map
	if err := json.Unmarshal(body, &res); err != nil {
		ch <- types.HealthResult{Err: err}
		return
	}
	ch <- types.HealthResult{Response: res}
}
