# go-queue - Documentation

> Back to [README](../README.md)

## Prerequisites

- Go 1.23 or higher
- No external services (standard library only)

## Installation

### Using go get

```bash
go get github.com/pardnchiu/go-queue
```

```go
import "github.com/pardnchiu/go-queue/core"
```

> The `core` sub-package path ships in the first release after v1.1.3; until that tag exists, use `go get github.com/pardnchiu/go-queue@develop`.

### From Source

```bash
git clone https://github.com/pardnchiu/go-queue.git
cd go-queue
go test -race ./...
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

	"github.com/pardnchiu/go-queue/core"
)

func main() {
	q := core.New(&core.Config{Workers: 4})

	ctx := context.Background()
	if err := q.Start(ctx); err != nil {
		log.Fatal(err)
	}

	for i := range 3 {
		id, err := q.Enqueue(ctx, "", func(ctx context.Context) error {
			fmt.Println("task", i, "running")
			return nil
		})
		if err != nil {
			log.Printf("enqueue: %v", err)
			continue
		}
		fmt.Println("enqueued", id)
	}

	// Drain the queue, waiting at most 10 seconds
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := q.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
```

### Advanced: Presets, Retry, and Callbacks

```go
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/pardnchiu/go-queue/core"
)

func main() {
	// Enable Debug level to see task.promoted / task.timeout_triggered events
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	q := core.New(&core.Config{
		Workers: 8,
		Size:    1024,
		Timeout: 30 * time.Second,
		Preset: map[string]core.PresetConfig{
			"payment": {Priority: core.PriorityHigh},
			"email":   {Priority: core.PriorityNormal},
			"report":  {Priority: core.PriorityLow, Timeout: 60 * time.Second},
		},
	})

	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	if err := q.Start(ctx); err != nil {
		log.Fatal(err)
	}

	// Retry up to 5 times on failure, fire the callback on success
	_, err := q.Enqueue(ctx, "payment", func(ctx context.Context) error {
		return charge(ctx)
	},
		core.WithTaskID("order-1001"),
		core.WithRetry(5),
		core.WithCallback(func(id string) {
			slog.Info("charged", "id", id)
		}),
	)
	if err != nil {
		log.Printf("enqueue payment: %v", err)
	}

	// WithTimeout overrides the preset-derived timeout (no 15-120s clamp)
	_, err = q.Enqueue(ctx, "report", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return nil
		}
	}, core.WithTimeout(5*time.Minute))
	if err != nil {
		log.Printf("enqueue report: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := q.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

var errDeclined = errors.New("card declined")

func charge(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return errDeclined
}
```

### Behavior Notes

| Scenario | Behavior |
|----------|----------|
| `Enqueue` before `Start` | Allowed; tasks wait in the heap and run by priority once `Start` is called |
| Task ignores `ctx.Done()` | After the timeout the worker reports an error and moves on, but the task goroutine keeps running until it returns |
| Task fails and retries during `Shutdown` | The state is already Closed, so re-enqueue fails, `task.retry_failed` is logged, and the task is dropped |
| The ctx passed to `Start` is cancelled | Every task ctx is cancelled; workers keep running and subsequent tasks fail immediately with `context canceled` |
| `Shutdown` exceeds its deadline | Cancels every task ctx and returns an error; remaining tasks run with a cancelled ctx |
| Callback | Runs in its own goroutine only on success; failures, timeouts, and exhausted retries never trigger it |

## API Reference

### Construction and Lifecycle

| Function | Signature | Description |
|----------|-----------|-------------|
| `New` | `func New(config *Config) *Queue` | Creates a queue; `config` may be `nil`, and zero-value fields take defaults |
| `Start` | `func (q *Queue) Start(ctx context.Context) error` | Launches `Workers` workers; `ctx` is the parent of every task ctx |
| `Enqueue` | `func (q *Queue) Enqueue(ctx context.Context, presetName string, action func(ctx context.Context) error, options ...EnqueueOption) (string, error)` | Pushes a task and returns its ID; `ctx` is only checked before enqueueing and is not passed to the task |
| `Shutdown` | `func (q *Queue) Shutdown(ctx context.Context) error` | Rejects new tasks and waits for workers to drain the queue; returns an error early when `ctx` expires. Repeated calls return `nil` |

### Config

```go
type Config struct {
	Workers int
	Size    int
	Timeout time.Duration
	Preset  map[string]PresetConfig
}

type PresetConfig struct {
	Priority Priority
	Timeout  time.Duration
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `Workers` | `runtime.NumCPU() * 2` | Number of concurrent workers |
| `Size` | `Workers * 64` | Heap capacity (retries included); `Enqueue` errors when full |
| `Timeout` | `30 * time.Second` | Global base timeout; also sets the promotion thresholds |
| `Preset` | empty map | Preset name → `PresetConfig` |
| `PresetConfig.Priority` | `PriorityImmediate` (zero value) | Applies when unset or when the preset name is unknown, including `""` |
| `PresetConfig.Timeout` | `0` | When `> 0`, replaces `Config.Timeout` as this preset's base timeout |

### Priority

Lower values run first; ties break by enqueue (or re-enqueue) time.

| Constant | Value | Timeout Multiplier |
|----------|-------|--------------------|
| `PriorityImmediate` | `0` | base ÷ 4 |
| `PriorityHigh` | `1` | base ÷ 2 |
| `PriorityRetry` | `2` | base ÷ 2 |
| `PriorityNormal` | `3` | base × 1 |
| `PriorityLow` | `4` | base × 2 |

**Timeout:** `clamp(base × multiplier, 15s, 120s)`, where base is `PresetConfig.Timeout` when `> 0`, otherwise `Config.Timeout`. The multiplier comes from the preset's priority; promotion after enqueue does not change the timeout. `WithTimeout` replaces the result and skips the clamp.

**Promotion:** checked across the whole heap each time a worker pops a task.

| From | To | Wait Threshold |
|------|----|----------------|
| `PriorityLow` | `PriorityNormal` | `clamp(Config.Timeout, 30s, 120s)` |
| `PriorityNormal` | `PriorityHigh` | `clamp(Config.Timeout × 2, 30s, 120s)` |

### EnqueueOption

| Option | Signature | Description |
|--------|-----------|-------------|
| `WithTaskID` | `func WithTaskID(id string) EnqueueOption` | Custom task ID; defaults to a UUID v4 |
| `WithTimeout` | `func WithTimeout(d time.Duration) EnqueueOption` | Overrides this task's timeout |
| `WithCallback` | `func WithCallback(fn func(id string)) EnqueueOption` | Called with the task ID after success |
| `WithRetry` | `func WithRetry(retryMax ...int) EnqueueOption` | Enables retry; defaults to 3 retries. Total runs = `retryMax + 1` |

Retries re-enter immediately at `PriorityRetry` with no backoff delay.

### Errors

| Source | Message |
|--------|---------|
| `Start` | `queue already started`, `queue already closed` |
| `Enqueue` | `ctx.Err()`, `enqueue failed: staging queue is full`, `enqueue failed: staging queue is closed` |
| `Shutdown` | `shutdown timeout: N tasks remaining` |
| Task execution | `task timeout after <d>`, `panic: <value>` |

### slog Events

| Event | Level | Trigger |
|-------|-------|---------|
| `task.completed` | Info | Task succeeded |
| `task.retrying` | Warn | Task failed with retries left |
| `task.failed` | Error | Task without retry failed |
| `task.exhausted` | Error | Retries exhausted |
| `task.retry_failed` | Error | Re-enqueue failed (queue full or closed) |
| `task.promoted` | Debug | Task promoted |
| `task.timeout_triggered` | Debug | Task timed out |

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
