package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/imhaqer/order-processing-service/internal/handler"
	"github.com/imhaqer/order-processing-service/internal/storage"
	"github.com/imhaqer/order-processing-service/internal/worker"
)

func main() {
	// Configuration
	const (
		numWorkers = 5
		queueSize  = 100
		serverPort = ":8080"
	)

	log.Println("Starting Order Processing Service...")

	// Initialize storage
	store := storage.NewMemoryStorage()
	log.Println("Initialized in-memory storage")

	// Initialize worker pool
	pool := worker.NewPool(numWorkers, queueSize, store)
	pool.Start()

	// Initialize handlers
	orderHandler := handlers.NewOrderHandler(store, pool)

	// Setup routes
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			orderHandler.CreateOrder(w, r)
		case http.MethodGet:
			orderHandler.GetAllOrders(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Get order by ID
	http.HandleFunc("/orders/", orderHandler.GetOrder)

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("status: OK"))
	})

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		
		log.Println("Shutting down gracefully...")
		pool.Close()
		os.Exit(0)
	}()

	// Start server
	log.Printf("Server listening on %s", serverPort)
	log.Println("Endpoints:")
	log.Println("  POST   /orders      - Create new order")
	log.Println("  GET    /orders      - Get all orders")
	log.Println("  GET    /orders/{id} - Get order by ID")
	log.Println("  GET    /health      - Health check")
	
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
