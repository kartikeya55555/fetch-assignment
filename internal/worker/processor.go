package worker

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kartikeya55555/fetch-assignment/internal/aws"
	"github.com/kartikeya55555/fetch-assignment/internal/models"
	"github.com/kartikeya55555/fetch-assignment/internal/receipt"
)

type Processor interface {
	ProcessMessage(msg *sqs.Message) error
}

type processor struct {
	sqsClient aws.SQSClient
	snsClient aws.SNSClient
	service   receipt.ReceiptService
}

func NewProcessor(
	sqsClient aws.SQSClient,
	snsClient aws.SNSClient,
	service receipt.ReceiptService,
) Processor {
	return &processor{
		sqsClient: sqsClient,
		snsClient: snsClient,
		service:   service,
	}
}

func (p *processor) ProcessMessage(msg *sqs.Message) error {
	// Log the raw message body at the start
	log.Printf("[Worker] Start ProcessMessage - raw message: %s\n", *msg.Body)

	// Attempt to unmarshal into our Receipt struct
	var r models.Receipt
	err := json.Unmarshal([]byte(*msg.Body), &r)
	if err != nil {
		return fmt.Errorf("[Worker] failed to unmarshal receipt: %v", err)
	}

	// Log what we got from unmarshaling
	log.Printf("[Worker] Unmarshalled Receipt => ID=%s, Status=%s, Retailer=%s, Points=%d\n",
		r.ID, r.Status, r.Retailer, r.Points)

	// Call the service to do the heavy-lifting (validation, points, store update)
	log.Printf("[Worker] Calling service.ProcessReceipt for ID=%s\n", r.ID)
	receiptID, err := p.service.ProcessReceipt(&r)
	if err != nil {
		log.Printf("[Worker] Receipt FAILED: %v\n", err)
		// Publish failure to SNS
		failMsg := fmt.Sprintf("Receipt %s failed: %v", r.ID, err)
		if snsErr := p.snsClient.Publish(failMsg); snsErr != nil {
			log.Printf("[Worker] Failed to publish failure message to SNS: %v\n", snsErr)
		}
		return err
	}

	// If we get here, ProcessReceipt succeeded => should have updated store, set status=COMPLETED
	log.Printf("[Worker] Receipt processed successfully with ID: %s (Status now=%s)\n", receiptID, r.Status)

	// Publish success to SNS (optional error check)
	successMsg := fmt.Sprintf("Receipt %s processed successfully.", receiptID)
	if snsErr := p.snsClient.Publish(successMsg); snsErr != nil {
		log.Printf("[Worker] Failed to publish success message to SNS: %v\n", snsErr)
	} else {
		log.Printf("[Worker] Successfully published success message to SNS for ID=%s\n", receiptID)
	}

	// Return nil => message was processed successfully
	return nil
}
