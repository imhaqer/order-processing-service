package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/imhaqer/order-processing-service/internal/models"
	"github.com/imhaqer/order-processing-service/internal/storage"
	"github.com/imhaqer/order-processing-service/internal/worker"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	storage *storage.MemoryStorage
	pool    *worker.Pool
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(storage *storage.MemoryStorage, pool *worker.Pool) *OrderHandler {
	return &OrderHandler{
		storage: storage,
		pool:    pool,
	}
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.OrderRequest  // Define a struct for the request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.CustomerID == "" || req.RestaurantID == "" || len(req.Items) == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create order
	order := &models.Order{
		ID:           fmt.Sprintf("ORD-%d", time.Now().UnixNano()),
		CustomerID:   req.CustomerID,
		RestaurantID: req.RestaurantID,
		Items:        req.Items,
		Status:       models.StatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to storage
	h.storage.Save(order)

	// Submit to worker pool for processing
	h.pool.Submit(order.ID)

	log.Printf("Created order %s for customer %s", order.ID, order.CustomerID)

	// Return created order
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(order)
}

// GetOrder handles GET /orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract order ID from URL path
	orderID := r.URL.Path[len("/orders/"):]
	if orderID == "" {
		http.Error(w, "Order ID required", http.StatusBadRequest)
		return
	}

	// Retrieve order
	order, err := h.storage.Get(orderID)
	if err != nil {
		if err == storage.ErrOrderNotFound {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return order
	w.Header().Set("Content-Type", "application/json") 
	json.NewEncoder(w).Encode(order)
}

// GetAllOrders handles GET /orders
func (h *OrderHandler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve all orders
	orders := h.storage.GetAll()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
