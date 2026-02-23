package worker

import (
	"log"
	"time"
	"sync"

	"github.com/imhaqer/order-processing-service/internal/models"
	"github.com/imhaqer/order-processing-service/internal/storage"
)

// manages a pool of workers processing orders concurrently
type Pool struct {
	orderQueue 	chan string
	storage    	*storage.MemoryStorage
	numWorkers 	int
	wg 			sync.WaitGroup // zero-value ready to use
}

func NewPool(numWorkers int, queueSize int, storage *storage.MemoryStorage) *Pool {
	return &Pool{
		orderQueue: 	make(chan string, queueSize),
		storage:    	storage,
		numWorkers: 	numWorkers,
	}
}


func (p *Pool) Start() {
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		workerID := i
		go func() {
			defer p.wg.Done()  // mark done when exits
			p.worker(workerID)
		}()
	}
	log.Printf("Started %d workers", p.numWorkers)
}

func (p *Pool) worker(id int) {
	for orderID := range p.orderQueue {
		log.Printf("Worker %d: Processing order %s", id, orderID)
		
		// Update status to processing
		if err := p.storage.UpdateStatus(orderID, models.StatusProcessing); err != nil {
			log.Printf("Worker %d: Error updating order %s: %v", id, orderID, err)
			continue
		}
		
		// Simulate order processing (external API calls, validation, etc.)
		if err := p.processOrder(orderID); err != nil {
			log.Printf("Worker %d: Failed to process order %s: %v", id, orderID, err)
			p.storage.UpdateStatus(orderID, models.StatusFailed)
			continue
		}
		
		// Mark as completed
		if err := p.storage.UpdateStatus(orderID, models.StatusCompleted); err != nil {
			log.Printf("Worker %d: Error marking order %s as completed: %v", id, orderID, err)
		}
		
		log.Printf("Worker %d: Completed order %s", id, orderID)
	}
}

// processOrder simulates external service calls and order processing
func (p *Pool) processOrder(orderID string) error {
	order, err := p.storage.Get(orderID)
	if err != nil {
		return err
	}
	
	// Simulate restaurant confirmation (1-2 seconds)
	time.Sleep(time.Duration(1+len(order.Items)%2) * time.Second)
	
	// Simulate payment processing (0.5 seconds)
	time.Sleep(500 * time.Millisecond)
	
	// Simulate courier assignment (1 second)
	time.Sleep(time.Second)
	
	// 10% random failure rate for demo purposes
	/*if time.Now().UnixNano()%10 == 0 {
		return fmt.Errorf("simulated processing failure")
	}*/
	
	return nil
}

// Submit adds an order to the processing queue
func (p *Pool) Submit(orderID string) {
	p.orderQueue <- orderID
}

// Close shuts down the worker pool
func (p *Pool) Close() {
	log.Println("Initiating graceful shutdown...")

	close(p.orderQueue)
	p.wg.Wait()
	p.wg.Wait()
	log.Println("All workers stopped")
}
