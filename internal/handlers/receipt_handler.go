package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kartikeya55555/fetch-assignment/internal/aws"
	"github.com/kartikeya55555/fetch-assignment/internal/models"
	"github.com/kartikeya55555/fetch-assignment/internal/receipt"
)

type ReceiptHandler interface {
	QueueReceipt(c *gin.Context)
	GetReceiptPoints(c *gin.Context)
}

type receiptHandler struct {
	service   receipt.ReceiptService
	sqsClient aws.SQSClient
}

func NewReceiptHandler(service receipt.ReceiptService, sqs aws.SQSClient) ReceiptHandler {
	return &receiptHandler{service: service, sqsClient: sqs}
}

// POST /receipts/process
func (h *receiptHandler) QueueReceipt(c *gin.Context) {
	var r models.Receipt
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Assign ID + set to PENDING
	r.ID = uuid.NewString()
	r.Status = "PENDING"

	// store in-memory so GET can see "PENDING"
	if err := h.service.StorePendingReceipt(&r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store pending receipt"})
		return
	}

	// enqueue message for the worker
	if err := h.sqsClient.SendMessage(r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue receipt: " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"id": r.ID, "status": "Receipt queued"})
}

// GET /receipts/:id/points
func (h *receiptHandler) GetReceiptPoints(c *gin.Context) {
	id := c.Param("id")

	// Log what ID we're trying to retrieve
	log.Printf("[GetReceiptPoints] Received request for ID: %s", id)

	rec, err := h.service.GetReceipt(id)
	log.Println("receipt from get ", rec)
	if err != nil {
		// Log the error that occurred while trying to get the receipt
		log.Printf("[GetReceiptPoints] Could not find receipt with ID=%s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Log the receipt's status and any error message before deciding the response
	log.Printf("[GetReceiptPoints] Fetched Receipt => ID=%s, Status=%s, Points=%d, ErrorMessage=%q",
		rec.ID, rec.Status, rec.Points, rec.ErrorMessage)

	// Switch on status
	switch rec.Status {
	case "PENDING":
		log.Printf("[GetReceiptPoints] Receipt is still PENDING for ID=%s", rec.ID)
		c.JSON(http.StatusOK, gin.H{
			"status":       "PENDING",
			"message":      "Still processing. Please try again later.",
			"errorMessage": rec.ErrorMessage,
			"pointsSoFar":  rec.Points,
		})
	case "FAILED":
		log.Printf("[GetReceiptPoints] Receipt has FAILED for ID=%s, Reason=%s", rec.ID, rec.ErrorMessage)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":       "FAILED",
			"errorMessage": rec.ErrorMessage,
		})
	case "COMPLETED":
		log.Printf("[GetReceiptPoints] Receipt is COMPLETED for ID=%s, Points=%d", rec.ID, rec.Points)
		c.JSON(http.StatusOK, gin.H{
			"status": "COMPLETED",
			"points": rec.Points,
		})
	default:
		log.Printf("[GetReceiptPoints] Receipt has UNKNOWN status for ID=%s: %s", rec.ID, rec.Status)
		c.JSON(http.StatusOK, gin.H{
			"status": "UNKNOWN",
		})
	}
}
