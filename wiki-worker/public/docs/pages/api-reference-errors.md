# Errors

This page lists every error message go-queue returns or logs, what triggers it, and how to match it.

## Error Messages

| Source | Message | Trigger |
|---|---|---|
| `Start` | `queue already started` | Already Running |
| `Start` | `queue already closed` | Already Closed |
| `Enqueue` | `ctx.Err()` as-is | The passed ctx is cancelled or expired |
| `Enqueue` | `enqueue failed: staging queue is full` | Heap length reached `Config.Size` |
| `Enqueue` | `enqueue failed: staging queue is closed` | `Shutdown` was called |
| `Shutdown` | `shutdown timeout: N tasks remaining` | `ctx` expired before the heap drained |
| Task (logged) | `task timeout after <d>` | Task exceeded its timeout |
| Task (logged) | `panic: <value>` | Task function panicked |
| Task (logged) | `context canceled` | The `Start` ctx was cancelled, or the `Shutdown` deadline expired while the task was running or still queued |

## Matching Errors

ctx errors from `Enqueue` are returned unwrapped, so `errors.Is(err, context.Canceled)` and `errors.Is(err, context.DeadlineExceeded)` work. The full / closed errors have no exported sentinel and can only be matched by message:

```go
if _, err := q.Enqueue(ctx, "email", sendEmail); err != nil {
	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		// caller gave up
	case strings.Contains(err.Error(), "staging queue is full"):
		// back off and retry later
	case strings.Contains(err.Error(), "staging queue is closed"):
		// queue is shutting down
	}
}
```

Task-side errors never return to the caller; they appear in [Log Events](/api-reference-events).
