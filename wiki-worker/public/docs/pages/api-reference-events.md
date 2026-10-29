# Log Events

This page lists every `slog` event go-queue emits, its level and fields, and how to enable the Debug events.

## slog Events

| Event | Level | Fields |
|---|---|---|
| `task.completed` | Info | `id`, `preset`, `retry_times`, `elapsed_ms` |
| `task.retrying` | Warn | `id`, `preset`, `retry_times`, `retry_max`, `error`, `elapsed_ms` |
| `task.failed` | Error | `id`, `preset`, `error`, `elapsed_ms` |
| `task.exhausted` | Error | `id`, `preset`, `retry_times`, `retry_max`, `error`, `elapsed_ms` |
| `task.retry_failed` | Error | `id`, `preset`, `retry_times`, `retry_max`, `error`, `retry_error` |
| `task.promoted` | Debug | `id`, `from`, `to` |
| `task.timeout_triggered` | Debug | `id`, `preset`, `timeout` |

`task.failed` is logged when retry is disabled; `task.exhausted` when retry is enabled and the limit is reached.

## Enabling Debug Events

go-queue logs through the default `log/slog` logger and has no logger setting of its own. To see `task.promoted` and `task.timeout_triggered`, set the default logger to Debug:

```go
slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
```
