# API Reference

This page lists the `Queue` methods and the exported types of `github.com/pardnchiu/go-queue/core`.

## Queue

```go
func New(config *Config) *Queue
func (q *Queue) Start(ctx context.Context) error
func (q *Queue) Enqueue(ctx context.Context, presetName string, action func(ctx context.Context) error, options ...EnqueueOption) (string, error)
func (q *Queue) Shutdown(ctx context.Context) error
```

| Method | Description |
|---|---|
| `New` | Creates a queue and merges defaults; `config` may be `nil`. The queue starts in the Created state |
| `Start` | Launches `Workers` workers. `ctx` is the parent of every task ctx; cancelling it does not stop workers, but later tasks receive an already-cancelled ctx |
| `Enqueue` | Pushes a task and returns its ID. `ctx` is only checked for cancellation before enqueueing and is not passed to the task; the task gets a timeout ctx derived from the `Start` ctx |
| `Shutdown` | Moves to Closed, wakes every worker, and waits for the heap to drain. When `ctx` expires it cancels every task ctx and returns an error; repeated calls return `nil` |

## Config and PresetConfig

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

Fields and defaults: [Configuration](/configuration) and [Presets](/configuration-presets).

## Priority

```go
type Priority int

const (
	PriorityImmediate Priority = iota
	PriorityHigh
	PriorityRetry
	PriorityNormal
	PriorityLow
)
```

Lower values run first. See [Priority Levels](/core-concepts-priority) and [Timeout Derivation](/core-concepts-timeout).

Other reference pages: [Enqueue Options](/api-reference-options), [Errors](/api-reference-errors), [Log Events](/api-reference-events), [Removed API](/removed-api).
