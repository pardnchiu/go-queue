# 日誌事件

本頁列出 go-queue 發出的每個 `slog` 事件、其等級與欄位，以及如何開啟 Debug 事件。

## slog 事件

| 事件 | 等級 | 欄位 |
|---|---|---|
| `task.completed` | Info | `id`、`preset`、`retry_times`、`elapsed_ms` |
| `task.retrying` | Warn | `id`、`preset`、`retry_times`、`retry_max`、`error`、`elapsed_ms` |
| `task.failed` | Error | `id`、`preset`、`error`、`elapsed_ms` |
| `task.exhausted` | Error | `id`、`preset`、`retry_times`、`retry_max`、`error`、`elapsed_ms` |
| `task.retry_failed` | Error | `id`、`preset`、`retry_times`、`retry_max`、`error`、`retry_error` |
| `task.promoted` | Debug | `id`、`from`、`to` |
| `task.timeout_triggered` | Debug | `id`、`preset`、`timeout` |

未啟用重試時記錄 `task.failed`；啟用重試且達到上限時記錄 `task.exhausted`。

## 開啟 Debug 事件

go-queue 透過預設的 `log/slog` logger 輸出，本身沒有 logger 設定。要看到 `task.promoted` 與 `task.timeout_triggered`，把預設 logger 設為 Debug：

```go
slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
```
