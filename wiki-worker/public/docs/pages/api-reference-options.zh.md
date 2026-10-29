# Enqueue 選項

本頁列出四個 `EnqueueOption` 函式，用來調整單一任務的 ID、逾時、成功 Callback 與重試。

## EnqueueOption

```go
type EnqueueOption func(*enqueueConfig)

func WithTaskID(id string) EnqueueOption
func WithTimeout(d time.Duration) EnqueueOption
func WithCallback(fn func(id string)) EnqueueOption
func WithRetry(retryMax ...int) EnqueueOption
```

| 選項 | 未使用時 | 使用時 |
|---|---|---|
| `WithTaskID` | 產生 UUID v4 | 使用指定 ID；不檢查重複 |
| `WithTimeout` | 由 Preset 推算並限制在 15–120 秒 | 直接使用 `d` |
| `WithCallback` | 無 | 成功後以新 goroutine 帶任務 ID 呼叫 |
| `WithRetry` | 失敗一次即結束 | 不帶參數最多重試 3 次；`WithRetry(n)` 最多 `n` 次 |

選項依序套用；同一選項傳入多次時，以最後一個為準。

## 常見需求

| 需求 | 選項 |
|---|---|
| 自訂任務 ID | `core.WithTaskID("order-1001")` |
| 單次逾時（不受 15–120 秒限制） | `core.WithTimeout(5 * time.Minute)` |
| 成功時通知 | `core.WithCallback(func(id string) { ... })` |
| 失敗時重試 | `core.WithRetry()` 或 `core.WithRetry(5)` |

```go
id, err := q.Enqueue(ctx, "email", sendEmail,
	core.WithTaskID("order-1001"),
	core.WithRetry(5),
	core.WithCallback(func(id string) { log.Println("done", id) }),
)
```

優先權無法逐任務指定，一律來自 Preset，見 [Preset](/zh/configuration-presets)。重試行為細節見 [執行與重試](/zh/core-concepts-execution)。
