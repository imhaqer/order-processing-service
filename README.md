# order-processing-service
 **Work in Progress ...**

A concurrent order processing system built in Go, demonstrating goroutines, channels, and thread-safe storage patterns for handling delivery orders at scale.

## Features

- **Concurrent Processing**: Worker pool with 5 goroutines processing orders in parallel
- **Thread-Safe Storage**: In-memory storage with mutex locks for safe concurrent access
- **Asynchronous Queue**: Buffered channel (capacity: 100) for order queue management
- **RESTful API**: HTTP endpoints for order creation and status tracking
