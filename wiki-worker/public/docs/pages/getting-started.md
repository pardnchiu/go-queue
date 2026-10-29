# Getting Started

This page covers installing go-queue, starting a worker pool, enqueueing a first task, and shutting the queue down within a deadline.

## Prerequisites

| Item | Requirement |
|---|---|
| Go | 1.23 or higher (`go.mod` declares `go 1.23`) |
| External dependencies | None; standard library only |

## Installation

```bash
go get github.com/pardnchiu/go-queue
```

```go
import "github.com/pardnchiu/go-queue/core"
```

The public API lives in the `core` sub-package. This path ships in the first release after v1.1.3; until that tag exists, use `go get github.com/pardnchiu/go-queue@develop`.

## Your First Queue

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

Each task emits one `task.completed` `slog` record with the task ID, preset name, and elapsed time.

## Three Lifecycle Calls

| Step | Call | Key Point |
|---|---|---|
| Create | `core.New(cfg)` | `cfg` may be `nil`; zero-value fields take defaults |
| Start | `q.Start(ctx)` | Launches `Workers` goroutines; cancelling `ctx` cancels every running task's ctx |
| Shut down | `q.Shutdown(ctx)` | Rejects new tasks, then waits for workers to drain the heap; returns an error when `ctx` expires |

`Enqueue` also works before `Start`: tasks wait in the heap and run by priority once `Start` is called.

## Next Steps

- Priority levels and promotion → [Priority Levels](/core-concepts-priority), [Anti-Starvation Promotion](/core-concepts-promotion)
- How long a task may run → [Timeout Derivation](/core-concepts-timeout)
- Tuning `Config` and presets → [Configuration](/configuration), [Presets](/configuration-presets)
- Retries, callbacks, and custom IDs → [Enqueue Options](/api-reference-options)
- Full signatures and error messages → [API Reference](/api-reference), [Errors](/api-reference-errors)
