package background

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/config"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/database"
	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/types"
)

func StartWorker() {
	workerID := os.Getenv("INSTANCE_ID")
	if workerID == "" {
		workerID = fmt.Sprintf("worker-%d", time.Now().UnixNano())
	}
	processingQueue := "payments_processing:" + workerID

	concurrency := 16

	for i := 0; i < concurrency; i++ {
		go func(workerNum int) {
			for {
				res, err := database.RedisClient.RPopLPush(context.Background(), config.RedisQueueKey, processingQueue).Result()
				if err != nil {
					if err.Error() != "redis: nil" {
						fmt.Println("Error moving payment to processing queue:", err)
					}
					time.Sleep(100 * time.Millisecond)
					continue
				}

				var payment types.PaymentRequest

				if err := json.Unmarshal([]byte(res), &payment); err != nil {
					fmt.Printf("[Worker %s-%d] Failed to unmarshal payment: %v\n", workerID, workerNum, err)
					continue
				}

				if err := ProccessPayments(payment); err != nil {
					database.RedisClient.LPush(database.RedisCtx, "payments_pending", res)
				} else {
					fmt.Printf("[Worker %s-%d] Payment processed: %s\n", workerID, workerNum, payment.CorrelationId)
				}
			}
		}(i)
	}
}
