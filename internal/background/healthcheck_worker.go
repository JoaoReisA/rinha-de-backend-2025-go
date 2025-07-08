package background

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/config"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type healthStatusCache struct {
	mu                      sync.RWMutex
	Status                  fiber.Map
	BestPaymentProcessorUrl string
	err                     error
}
type healthResult struct {
	response fiber.Map
	err      error
}

var HealthCache = healthStatusCache{
	Status:                  make(fiber.Map),
	BestPaymentProcessorUrl: config.PaymentProcessorUrlDefault,
	err:                     nil,
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
			HealthCache.err = err
			continue // Skip this update and wait for the next tick
		}

		// Safely write the new status to the global cache
		HealthCache.mu.Lock()
		HealthCache.Status = newStatus
		HealthCache.BestPaymentProcessorUrl, HealthCache.err = decideBestProcessor(newStatus)
		HealthCache.mu.Unlock()

		log.Println("Worker: Health status cache updated successfully.")
	}
}

func decideBestProcessor(m fiber.Map) (s string, err error) {
	defaultProcessorData, ok := m["default"].(fiber.Map)
	if !ok {
		log.Println("Error: 'default' health status is not in the expected format.")
		return
	}

	isDefaultFailing, ok := defaultProcessorData["failing"].(bool)
	if ok && isDefaultFailing {
		log.Println("Default processor is failing!")
		return config.PaymentProcessorUrlFallback, nil
	}
	fallbackProcessorData, ok := m["fallback"].(fiber.Map)
	if !ok {
		log.Println("Error: 'default' health status is not in the expected format.")
		return
	}
	isFallbackFailing, ok := fallbackProcessorData["failing"].(bool)
	if ok && isFallbackFailing {
		return config.PaymentProcessorUrlDefault, nil
	}

	if isDefaultFailing && isFallbackFailing {
		return config.PaymentProcessorUrlDefault, utils.ErrAllPaymentsFailing
	}

	return config.PaymentProcessorUrlDefault, nil
}

func fetchAndProcessHealthChecks() (fiber.Map, error) {
	var wg sync.WaitGroup
	defaultChan := make(chan healthResult, 1)
	fallbackChan := make(chan healthResult, 1)

	wg.Add(2)
	go fetchHealth(config.PaymentProcessorUrlDefault+"/payments/service-health", &wg, defaultChan)
	go fetchHealth(config.PaymentProcessorUrlFallback+"/payments/service-health", &wg, fallbackChan)

	wg.Wait()
	close(defaultChan)
	close(fallbackChan)

	defaultResult := <-defaultChan
	fallbackResult := <-fallbackChan

	if defaultResult.err != nil {
		return nil, defaultResult.err
	}
	if fallbackResult.err != nil {
		return nil, fallbackResult.err
	}

	response := fiber.Map{
		"default":          defaultResult.response,
		"fallback":         fallbackResult.response,
		"last_checked_utc": time.Now().UTC(),
	}

	return response, nil
}
func fetchHealth(url string, wg *sync.WaitGroup, ch chan<- healthResult) {
	defer wg.Done()
	agent := fiber.Get(url)
	_, body, errs := agent.Bytes()
	if len(errs) > 0 {
		ch <- healthResult{err: errs[0]}
		return
	}
	var res fiber.Map
	if err := json.Unmarshal(body, &res); err != nil {
		ch <- healthResult{err: err}
		return
	}
	ch <- healthResult{response: res}
}
