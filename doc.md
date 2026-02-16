# go-queue - Documentation

> Back to [README](./README.md)

## Prerequisites

- Go 1.23 or higher

## Installation

### Using go get

```bash
go get github.com/pardnchiu/go-queue
```

### From Source

```bash
git clone https://github.com/pardnchiu/go-queue.git
cd go-queue
go test ./...
```

## Usage

### Basic

Create a queue, start workers, and submit tasks:

```go
package main

import (
	"context"
	"fmt"
	"log"

	goQueue "github.com/pardnchiu/go-queue"
)

func main() {
	q := goQueue.New(&goQueue.Config{
		Workers: 4,
	})

	ctx := context.Background()
	if err := q.Start(ctx); err != nil {
		log.Fatal(err)
	}

	id, err := q.Enqueue(ctx, "", func(ctx context.Context) error {
		fmt.Println("task executed")
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("task id:", id)

	if err := q.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
```

### Presets and Priority

Define task types with preset priority and timeout:

```go
q := goQueue.New(&goQueue.Config{
	Workers: 4,
	Timeout: 30 * time.Second,
	Preset: map[string]goQueue.PresetConfig{
		"critical": {Priority: goQueue.PriorityImmediate},
		"batch":    {Priority: goQueue.PriorityLow, Timeout: 60 * time.Second},
	},
})

ctx := context.Background()
q.Start(ctx)

// Enqueue with Immediate priority
q.Enqueue(ctx, "critical", func(ctx context.Context) error {
	// High priority task
	return nil
})

// Enqueue with Low priority, 60s timeout
q.Enqueue(ctx, "batch", func(ctx context.Context) error {
	// Low priority batch task
	return nil
})

q.Shutdown(ctx)
```

### Retry and Callback

Configure retry and completion callback for tasks:

```go
q := goQueue.New(&goQueue.Config{Workers: 2})
ctx := context.Background()
q.Start(ctx)

id, err := q.Enqueue(ctx, "", func(ctx context.Context) error {
	// Operation that may fail
	return doSomething()
},
	goQueue.WithRetry(3),                           // Retry up to 3 times
	goQueue.WithTaskID("my-task-001"),               // Custom task ID
	goQueue.WithTimeout(10*time.Second),             // Per-task timeout
	goQueue.WithCallback(func(id string) {           // Completion callback
		fmt.Printf("task %s completed\n", id)
	}),
)
if err != nil {
	log.Fatal(err)
}

q.Shutdown(ctx)
```

### Graceful Shutdown with Timeout

Limit shutdown wait time to prevent indefinite blocking:

```go
q := goQueue.New(&goQueue.Config{Workers: 4})
ctx := context.Background()
q.Start(ctx)

// ... submit tasks ...

shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := q.Shutdown(shutdownCtx); err != nil {
	fmt.Printf("shutdown timeout: %v\n", err)
}
```

## API Reference

### New

```go
func New(config *Config) *Queue
```

Create a new queue instance. Passing `nil` for config uses default values.

### Config

```go
type Config struct {
	Workers int                     // Number of workers, default CPU * 2
	Size    int                     // Queue capacity limit, default Workers * 64
	Timeout time.Duration           // Global task timeout, default 30s
	Preset  map[string]PresetConfig // Task type preset configurations
}
```

### PresetConfig

```go
type PresetConfig struct {
	Priority Priority      // Priority level, default PriorityNormal
	Timeout  time.Duration // Task timeout, 0 = auto-calculated by Priority
}
```

### Priority

```go
type Priority int

const (
	PriorityImmediate Priority = iota // Highest priority
	PriorityHigh                       // High priority
	PriorityRetry                      // Reserved for retried tasks
	PriorityNormal                     // Normal (default)
	PriorityLow                        // Low priority
)
```

Priority affects timeout calculation: Immediate = base/4, High = base/2, Normal = base, Low = base*2, clamped to the 15~120 second range.

### Queue.Start

```go
func (q *Queue) Start(ctx context.Context) error
```

Start the worker pool. Returns an error if called again or if the queue is already closed.

### Queue.Enqueue

```go
func (q *Queue) Enqueue(ctx context.Context, presetName string, action func(ctx context.Context) error, options ...EnqueueOption) (string, error)
```

Submit a task to the queue. `presetName` maps to a key in `Config.Preset`; an empty string uses defaults. Returns the task ID or an error (context cancelled, queue full, or queue closed).

### Queue.Shutdown

```go
func (q *Queue) Shutdown(ctx context.Context) error
```

Close the queue and wait for all workers to finish. Returns an error with the remaining task count if the context times out.

### EnqueueOption

| Function | Signature | Description |
|----------|-----------|-------------|
| `WithTaskID` | `func WithTaskID(id string) EnqueueOption` | Set a custom task ID; defaults to auto-generated UUID v4 |
| `WithTimeout` | `func WithTimeout(d time.Duration) EnqueueOption` | Override per-task timeout |
| `WithCallback` | `func WithCallback(fn func(id string)) EnqueueOption` | Callback invoked after successful task completion |
| `WithRetry` | `func WithRetry(retryMax ...int) EnqueueOption` | Enable retry with default max 3; optionally specify limit |

***

©️ 2025 [邱敬幃 Pardn Chiu](https://linkedin.com/in/pardnchiu)
