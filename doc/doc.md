# go-queue - Documentation

> Back to [README](../README.md)

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

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	goQueue "github.com/pardnchiu/go-queue/core"
)

func main() {
	q := goQueue.New(&goQueue.Config{
		Workers: 4,
	})

	ctx := context.Background()
	if err := q.Start(ctx); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := q.Shutdown(context.Background()); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	id, err := q.Enqueue(ctx, "", func(ctx context.Context) error {
		fmt.Println("task running")
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("enqueued:", id)
	time.Sleep(100 * time.Millisecond)
}
```

### Priority Presets

```go
q := goQueue.New(&goQueue.Config{
	Workers: 2,
	Timeout: 30 * time.Second,
	Preset: map[string]goQueue.PresetConfig{
		"urgent": {Priority: goQueue.PriorityImmediate},
		"batch":  {Priority: goQueue.PriorityLow, Timeout: 60 * time.Second},
	},
})

q.Start(ctx)

// high priority first
q.Enqueue(ctx, "urgent", func(ctx context.Context) error {
	// ...
	return nil
})

q.Enqueue(ctx, "batch", func(ctx context.Context) error {
	// ...
	return nil
})
```

### Enqueue Options

```go
id, err := q.Enqueue(ctx, "urgent", func(ctx context.Context) error {
	// business logic
	return nil
},
	goQueue.WithTaskID("order-123"),
	goQueue.WithTimeout(10*time.Second),
	goQueue.WithRetry(2),
	goQueue.WithCallback(func(id string) {
		fmt.Println("done:", id)
	}),
)
if err != nil {
	log.Fatal(err)
}
```

### Graceful Shutdown

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := q.Shutdown(shutdownCtx); err != nil {
	// timed out: tasks still remaining
	log.Printf("shutdown: %v", err)
}
```

## API Reference

### New

```go
func New(config *Config) *Queue
```

Creates a queue. Pass `nil` for defaults.

| Field | Type | Default | Description |
|------|------|------|------|
| `Workers` | `int` | `runtime.NumCPU() * 2` | Worker count |
| `Size` | `int` | `Workers * 64` | Pending queue capacity |
| `Timeout` | `time.Duration` | `30 * time.Second` | Base timeout |
| `Preset` | `map[string]PresetConfig` | empty | Named priority/timeout presets |

### PresetConfig

```go
type PresetConfig struct {
	Priority Priority
	Timeout  time.Duration
}
```

| Field | Description |
|------|------|
| `Priority` | Priority level; zero value is `PriorityImmediate` (iota 0) |
| `Timeout` | When `0`, derived from base `Timeout` by priority |

### Priority

| Constant | Value | Description |
|------|----|------|
| `PriorityImmediate` | 0 | Highest, run first |
| `PriorityHigh` | 1 | High priority |
| `PriorityRetry` | 2 | Retry tasks |
| `PriorityNormal` | 3 | Default |
| `PriorityLow` | 4 | Lowest; may be promoted after wait |

### Start

```go
func (q *Queue) Start(ctx context.Context) error
```

Starts the worker pool. Transitions only from `Created` to `Running`; returns error if already started or closed.

### Enqueue

```go
func (q *Queue) Enqueue(ctx context.Context, presetName string, action func(ctx context.Context) error, options ...EnqueueOption) (string, error)
```

Enqueues a task and returns its ID. Errors when the queue is closed, full, or `ctx` is canceled.

### EnqueueOption

| Option | Signature | Description |
|------|------|------|
| `WithTaskID` | `func WithTaskID(id string) EnqueueOption` | Custom task ID; auto UUID if omitted |
| `WithTimeout` | `func WithTimeout(d time.Duration) EnqueueOption` | Override per-task timeout |
| `WithCallback` | `func WithCallback(fn func(id string)) EnqueueOption` | Async callback after success |
| `WithRetry` | `func WithRetry(retryMax ...int) EnqueueOption` | Enable retries; default max 3 when arg omitted |

### Shutdown

```go
func (q *Queue) Shutdown(ctx context.Context) error
```

Closes the queue, drains pending work, and waits for workers. On `ctx` timeout, returns remaining task count. Idempotent.

### Timeout Derivation

Base is `Config.Timeout` (or preset override), then adjusted by priority and clamped to 15s–120s:

| Priority | Formula |
|----------|------|
| Immediate | `timeout / 4` |
| High / Retry | `timeout / 2` |
| Normal | `timeout` |
| Low | `timeout * 2` |

### Priority Promotion

| From | Wait | To |
|------|----------|------|
| Low | `clamp(Timeout, 30s, 120s)` | Normal |
| Normal | `clamp(Timeout*2, 30s, 120s)` | High |

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
