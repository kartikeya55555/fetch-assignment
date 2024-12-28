package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kartikeya55555/fetch-assignment/internal/aws"
	"github.com/kartikeya55555/fetch-assignment/internal/config"
	"github.com/kartikeya55555/fetch-assignment/internal/handlers"
	"github.com/kartikeya55555/fetch-assignment/internal/receipt"
	"github.com/kartikeya55555/fetch-assignment/internal/worker"
)

func main() {
	// 1) Load configuration
	cfg := config.LoadConfig()

	// 2) Set up AWS clients (SQS, SNS)
	sqsClient := aws.NewSQSClient(cfg.AWSRegion, cfg.AWSEndpoint, cfg.SQSQueueName)
	snsClient := aws.NewSNSClient(cfg.AWSRegion, cfg.AWSEndpoint, cfg.SNSTopicName)

	// Ensure the queue and topic exist
	if err := sqsClient.EnsureQueue(); err != nil {
		log.Fatalf("Failed to ensure queue: %v", err)
	}
	if err := snsClient.EnsureTopic(); err != nil {
		log.Fatalf("Failed to ensure topic: %v", err)
	}

	// 3) Create a single in-memory store shared by both API & worker
	store := receipt.NewInMemoryStore()
	calc := receipt.NewDefaultPointsCalculator()
	service := receipt.NewReceiptService(store, calc)

	// 4) Start the worker in a separate goroutine
	go func() {
		log.Println("Starting worker loop...")
		processor := worker.NewProcessor(sqsClient, snsClient, service)

		for {
			messages, err := sqsClient.GetMessages()
			if err != nil {
				log.Printf("Error receiving messages: %v\n", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, msg := range messages {
				if msg.Body == nil {
					continue
				}
				log.Printf("Worker received message: %s\n", *msg.Body)

				if err := processor.ProcessMessage(msg); err != nil {
					log.Printf("Error processing message: %v\n", err)
					// optionally skip DeleteMessage if you want to retry
					continue
				}
				// Delete message if processed successfully
				sqsClient.DeleteMessage(msg.ReceiptHandle)
				log.Println("Message processed and removed from queue.")
			}

			time.Sleep(2 * time.Second)
		}
	}()

	// 5) Set up the API (Gin)
	r := gin.Default()

	// Handler that knows how to queue receipts
	receiptHandler := handlers.NewReceiptHandler(service, sqsClient)

	r.POST("/receipts/process", receiptHandler.QueueReceipt)
	r.GET("/receipts/:id/points", receiptHandler.GetReceiptPoints)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "API is running"})
	})

	// 6) Run the API server in the main goroutine
	log.Printf("Starting unified API+Worker on port %s...\n", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run API server: %v", err)
	}
}
