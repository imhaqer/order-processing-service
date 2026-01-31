# CLAUDE.md

This document provides essential context for AI assistants working with the order-processing-service codebase.

## Project Overview

A concurrent order processing system built in Go, demonstrating goroutines, channels, and thread-safe storage patterns for handling delivery orders at scale. The service uses a worker pool pattern to process orders asynchronously.

**Module:** `github.com/imhaqer/order-processing-service`
**Go Version:** 1.25.4
**Status:** Work in Progress

## Quick Reference

```bash
# Build the service
go build -o server ./cmd/server

# Run the service
go run ./cmd/server/main.go

# Test endpoints
curl http://localhost:8080/health
curl -X POST http://localhost:8080/orders -H "Content-Type: application/json" \
  -d '{"customer_id":"cust-1","restaurant_id":"rest-1","items":["burger","fries"]}'
curl http://localhost:8080/orders
```

## Architecture

```
HTTP Request → OrderHandler → Worker Pool Queue → Workers → Storage
                                                      ↓
                                                Status Updates
```

### Core Components

| Component | Location | Purpose |
|-----------|----------|---------|
| Entry Point | `cmd/server/main.go` | HTTP server setup, routing, graceful shutdown |
| Handlers | `internal/handler/api.go` | HTTP request/response handling |
| Models | `internal/models/order.go` | Order and OrderStatus data structures |
| Storage | `internal/storage/memory.go` | Thread-safe in-memory order persistence |
| Worker Pool | `internal/worker/pool.go` | Concurrent order processing with goroutines |

### Configuration Constants (in main.go)

```go
numWorkers = 5      // Concurrent worker goroutines
queueSize  = 100    // Buffered channel capacity
serverPort = ":8080"
```

## Directory Structure

```
order-processing-service/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── handler/
│   │   └── api.go            # HTTP handlers (OrderHandler)
│   ├── models/
│   │   └── order.go          # Domain models (Order, OrderStatus, OrderRequest)
│   ├── storage/
│   │   └── memory.go         # In-memory storage with RWMutex
│   └── worker/
│       └── pool.go           # Worker pool implementation
├── go.mod
├── README.md
└── CLAUDE.md
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/orders` | Create new order |
| GET | `/orders` | Retrieve all orders |
| GET | `/orders/{id}` | Retrieve specific order by ID |
| GET | `/health` | Health check (returns "OK") |

### Order Request Format

```json
{
  "customer_id": "string (required)",
  "restaurant_id": "string (required)",
  "items": ["array", "of", "strings (required, non-empty)"]
}
```

### Order Response Format

```json
{
  "id": "ORD-{unix_nano}",
  "customer_id": "string",
  "restaurant_id": "string",
  "items": ["array"],
  "status": "pending|processing|completed|failed",
  "created_at": "RFC3339 timestamp",
  "updated_at": "RFC3339 timestamp"
}
```

## Code Conventions

### Naming

- **Packages:** lowercase, single-word (`handler`, `models`, `storage`, `worker`)
- **Types:** PascalCase (`Order`, `OrderStatus`, `Pool`, `MemoryStorage`)
- **Exported functions:** PascalCase (`NewPool`, `CreateOrder`)
- **Private functions:** camelCase (`processOrder`, `worker`)
- **Constants:** PascalCase with descriptive prefix (`StatusPending`, `StatusProcessing`)

### Error Handling

- Use sentinel errors for known error conditions (e.g., `ErrOrderNotFound`)
- Check errors with `if err != nil` pattern
- Log errors with `log.Printf` including context (worker ID, order ID)
- Return appropriate HTTP status codes (400, 404, 500)

### Concurrency Patterns

- **RWMutex:** Used in storage for thread-safe map access (RLock for reads, Lock for writes)
- **WaitGroup:** Used for graceful shutdown coordination
- **Buffered Channels:** Order queue with capacity for async processing
- **Goroutines:** Worker pool with configurable number of workers
- **Defer:** Always use `defer` for mutex unlocks and WaitGroup.Done()

### Struct Tags

Use JSON struct tags with snake_case naming:
```go
type Order struct {
    ID           string      `json:"id"`
    CustomerID   string      `json:"customer_id"`
    RestaurantID string      `json:"restaurant_id"`
}
```

### HTTP Patterns

- Set `Content-Type: application/json` header explicitly
- Use `http.StatusCreated` (201) for successful POST
- Use standard library `http.HandleFunc` for routing
- Validate request methods at handler level

## Dependencies

This project uses **only Go standard library** - no external dependencies:
- `net/http` - HTTP server
- `encoding/json` - JSON serialization
- `sync` - RWMutex, WaitGroup
- `log` - Logging
- `time` - Timestamps and delays
- `os`, `os/signal`, `syscall` - Graceful shutdown

## Important Implementation Details

### Order ID Generation
Order IDs are generated using Unix nanoseconds: `fmt.Sprintf("ORD-%d", time.Now().UnixNano())`

### Order Status Lifecycle
```
pending → processing → completed
                   ↘ failed
```

### Simulated Processing
The worker pool simulates external service calls:
- Restaurant confirmation: 1-2 seconds
- Payment processing: 500ms
- Courier assignment: 1 second
- 10% random failure rate for demonstration

### Graceful Shutdown
On SIGINT/SIGTERM:
1. Signal handler receives interrupt
2. Worker pool channel is closed
3. WaitGroup waits for all workers to finish
4. Application exits cleanly

## Common Tasks

### Adding a New Endpoint

1. Add handler method to `internal/handler/api.go`
2. Register route in `cmd/server/main.go` using `http.HandleFunc`
3. Follow existing patterns for JSON encoding/decoding

### Adding a New Order Status

1. Add constant to `internal/models/order.go`:
   ```go
   StatusNewStatus OrderStatus = "new_status"
   ```

### Modifying Worker Behavior

Edit `internal/worker/pool.go`:
- `worker()` - Main processing loop
- `processOrder()` - Individual order processing logic

### Adding Persistent Storage

Replace or extend `internal/storage/memory.go`:
1. Implement same interface methods: `Get`, `Save`, `UpdateStatus`, `GetAll`
2. Maintain thread-safety with appropriate locking

## Testing Notes

**Current state:** No test files exist in the codebase.

When adding tests:
- Place test files alongside source: `*_test.go`
- Use table-driven tests for handlers
- Mock storage for handler tests
- Test concurrent access for storage layer
- Run with: `go test ./...`

## Gotchas and Known Issues

1. **In-memory storage:** All data is lost on restart
2. **No input sanitization:** Order IDs extracted directly from URL path
3. **Duplicate WaitGroup.Wait():** Line 102-103 in `pool.go` calls `p.wg.Wait()` twice
4. **No request timeouts:** HTTP server has no configured timeouts
5. **No graceful HTTP shutdown:** Only worker pool is shut down gracefully

## File Locations Quick Reference

| What | Where |
|------|-------|
| Server config | `cmd/server/main.go:17-21` |
| Route setup | `cmd/server/main.go:37-53` |
| Order model | `internal/models/order.go:5-13` |
| Status constants | `internal/models/order.go:18-23` |
| Storage mutex | `internal/storage/memory.go:14-17` |
| Worker loop | `internal/worker/pool.go:42-66` |
| Order processing simulation | `internal/worker/pool.go:69-90` |
