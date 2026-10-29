# Enqueue Options

This page lists the four `EnqueueOption` functions that adjust a single task's ID, timeout, success callback, and retries.

## EnqueueOption

```go
type EnqueueOption func(*enqueueConfig)

func WithTaskID(id string) EnqueueOption
func WithTimeout(d time.Duration) EnqueueOption
func WithCallback(fn func(id string)) EnqueueOption
func WithRetry(retryMax ...int) EnqueueOption
```

| Option | Without It | With It |
|---|---|---|
| `WithTaskID` | Generates a UUID v4 | Uses the given ID; duplicates are not checked |
| `WithTimeout` | Derived from the preset and clamped to 15–120s | Uses `d` as-is |
| `WithCallback` | None | Called with the task ID in a new goroutine after success |
| `WithRetry` | Fails once and stops | Up to 3 retries with no argument; up to `n` retries with `WithRetry(n)` |

Options apply in order; when the same option is passed more than once, the last one wins.

## Common Needs

| Need | Option |
|---|---|
| Custom task ID | `core.WithTaskID("order-1001")` |
| One-off timeout (not clamped to 15–120s) | `core.WithTimeout(5 * time.Minute)` |
| Notify on success | `core.WithCallback(func(id string) { ... })` |
| Retry on failure | `core.WithRetry()` or `core.WithRetry(5)` |

```go
id, err := q.Enqueue(ctx, "email", sendEmail,
	core.WithTaskID("order-1001"),
	core.WithRetry(5),
	core.WithCallback(func(id string) { log.Println("done", id) }),
)
```

Priority cannot be set per task; it comes from the preset, see [Presets](/configuration-presets). Retry behavior in detail: [Execution and Retries](/core-concepts-execution).
