package receipt

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/kartikeya55555/fetch-assignment/internal/errors"
	"github.com/kartikeya55555/fetch-assignment/internal/models"
)

// ReceiptService interface
type ReceiptService interface {
	ProcessReceipt(r *models.Receipt) (string, error)
	GetPoints(id string) (int, error)
	GetReceipt(id string) (*models.Receipt, error)
	StorePendingReceipt(r *models.Receipt) error
}

type receiptService struct {
	store ReceiptStore
	calc  PointsCalculator
}

func NewReceiptService(store ReceiptStore, calc PointsCalculator) ReceiptService {
	return &receiptService{
		store: store,
		calc:  calc,
	}
}

// If the API wants to store a pending record:
func (s *receiptService) StorePendingReceipt(r *models.Receipt) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
	if r.Status == "" {
		r.Status = "PENDING"
	}
	return s.store.AddReceipt(r)
}

// Worker calls this to do heavy-lifting validations
func (s *receiptService) ProcessReceipt(r *models.Receipt) (string, error) {
	log.Printf("[Service] Starting ProcessReceipt for ID=%s (current status=%s)\n", r.ID, r.Status)
	issues := validateReceipt(r)
	if len(issues) > 0 {
		r.Status = "FAILED"
		r.ErrorMessage = strings.Join(issues, "; ")
		s.store.AddReceipt(r)
		log.Printf("[Service] Validation FAILED for ID=%s => %s\n", r.ID, r.ErrorMessage)
		return "", fmt.Errorf("[Service] receipt validation failed: %s", r.ErrorMessage)
	}

	r.Status = "COMPLETED"
	r.ErrorMessage = ""
	r.Points = s.calc.CalculatePoints(r)
	if err := s.store.AddReceipt(r); err != nil {
		log.Printf("[Service] Store AddReceipt error: %v\n", err)
		return "", err
	}

	log.Printf("[Service] Completed ProcessReceipt => ID=%s, Status=%s, Points=%d\n", r.ID, r.Status, r.Points)
	return r.ID, nil
}

func (s *receiptService) GetPoints(id string) (int, error) {
	rec, found := s.store.GetReceipt(id)
	if !found {
		return 0, errors.ErrReceiptNotExist
	}
	return rec.Points, nil
}

func (s *receiptService) GetReceipt(id string) (*models.Receipt, error) {
	rec, found := s.store.GetReceipt(id)
	if !found {
		return nil, errors.ErrReceiptNotExist
	}
	return rec, nil
}
