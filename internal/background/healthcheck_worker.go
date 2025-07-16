package background

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/config"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/database"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/types"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/utils"
	"github.com/gofiber/fiber/v2"
)

var HealthCache = types.HealthStatusCache{
	Status:                  make(fiber.Map),
	BestPaymentProcessorUrl: config.PaymentProcessorUrlDefault,
	Err:                     nil,
}

func StartWorkerHealthCheck() {
	log.Println("Initializing background worker...")
	const lockKey = "healthcheck-leader-lock"
	const lockTTL = 60 * time.Second

	for {
		acquired, err := database.RedisClient.SetNX(context.Background(), lockKey, "active", lockTTL).Result()
		if err != nil {
			log.Printf("Error acquiring Redis lock: %v. Retrying in 15 seconds.", err)
			time.Sleep(15 * time.Second)
			continue
		}

		if acquired {
			log.Println("Worker: Acquired leader lock. Starting health checks.")
			runHealthCheckAsLeader(lockKey, lockTTL)
		} else {
			log.Println("Worker: Could not acquire lock, another instance is leader. Waiting.")
			time.Sleep(lockTTL / 2)
		}
	}
}

func runHealthCheckAsLeader(lockKey string, lockTTL time.Duration) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	renewTicker := time.NewTicker(lockTTL / 2)
	defer renewTicker.Stop()

	log.Println("Leader: Running health checks...")

	for {
		select {
		case <-ticker.C:
			log.Println("Leader: Performing health check...")
			newStatus, err := fetchAndProcessHealthChecks()
			if err != nil {
				log.Printf("Leader: Error fetching health checks: %v", err)
				HealthCache.Err = err
				continue
			}

			HealthCache.Mu.Lock()
			HealthCache.Status = newStatus
			HealthCache.BestPaymentProcessorUrl, HealthCache.Err = decideBestProcessor(newStatus)
			HealthCache.Mu.Unlock()
		case <-renewTicker.C:
			log.Println("Leader: Renewing lock...")
			renewed, err := database.RedisClient.Expire(context.Background(), lockKey, lockTTL).Result()
			if err != nil || !renewed {
				log.Println("Leader: Failed to renew lock or lock expired. Relinquishing leadership.")
				return
			}
		}
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
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		log.Println("Error fetching HeathCheck", url, errs[0])

		ch <- types.HealthResult{Err: errs[0]}
		return
	}
	if statusCode == 429 {
		log.Println("Error fetching HeathCheck", url, statusCode)
	}
	if statusCode > 300 {
		log.Println("Error fetching HeathCheck", url, statusCode)
	}
	var res fiber.Map
	if err := json.Unmarshal(body, &res); err != nil {
		ch <- types.HealthResult{Err: err}
		return
	}
	ch <- types.HealthResult{Response: res}
}
