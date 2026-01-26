package storage

import (
	"errors"
	"sync"

	"github.com/imhaqer/order-processing-service/internal/models"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

type MemoryStorage struct {
	mu sync.RWMutex
	orders map[string]*models.Order
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		orders: make(map[string]*models.Order),
	}
}

// Get retrieves an order by ID (thread-safe)
func (s *MemoryStorage) Get(id string) (*models.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	order, exists := s.orders[id]
	if !exists {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s *MemoryStorage) Save(order *models.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.ID] = order
}

func (s *MemoryStorage) UpdateStatus(id string, status models.OrderStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	order, exists := s.orders[id]
	if !exists {
		return ErrOrderNotFound
	}
	
	order.Status = status
	return nil
}


func (s *MemoryStorage) GetAll() []*models.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	orders := make([]*models.Order, 0, len(s.orders))
	for _, order := range s.orders {
		orders = append(orders, order)
	}
	return orders
}