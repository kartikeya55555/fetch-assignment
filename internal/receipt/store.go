package receipt

import (
	"sync"

	"github.com/kartikeya55555/fetch-assignment/internal/models"
)

// ReceiptStore is the interface for storing/fetching receipts.
type ReceiptStore interface {
	AddReceipt(r *models.Receipt) error
	GetReceipt(id string) (*models.Receipt, bool)
}

type inMemoryStore struct {
	mu       sync.RWMutex
	receipts map[string]*models.Receipt
}

func NewInMemoryStore() ReceiptStore {
	return &inMemoryStore{
		receipts: make(map[string]*models.Receipt),
	}
}

func (s *inMemoryStore) AddReceipt(r *models.Receipt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.receipts[r.ID] = r
	return nil
}

func (s *inMemoryStore) GetReceipt(id string) (*models.Receipt, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, found := s.receipts[id]
	return rec, found
}
