# API 參考

本頁列出 `github.com/pardnchiu/go-queue/core` 的 `Queue` 方法與公開型別。

## Queue

```go
func New(config *Config) *Queue
func (q *Queue) Start(ctx context.Context) error
func (q *Queue) Enqueue(ctx context.Context, presetName string, action func(ctx context.Context) error, options ...EnqueueOption) (string, error)
func (q *Queue) Shutdown(ctx context.Context) error
```

| 方法 | 說明 |
|---|---|
| `New` | 建立佇列並合併預設值；`config` 可為 `nil`。佇列初始為 Created 狀態 |
| `Start` | 啟動 `Workers` 個 Worker。`ctx` 是所有任務 ctx 的父層；取消它不會停止 Worker，但之後的任務會拿到已取消的 ctx |
| `Enqueue` | 推入任務並回傳 ID。`ctx` 只在入列前檢查是否已取消，不會傳給任務；任務拿到的是由 `Start` ctx 衍生的逾時 ctx |
| `Shutdown` | 轉為 Closed、喚醒所有 Worker，並等待堆清空。`ctx` 到期時取消所有任務 ctx 並回傳錯誤；重複呼叫回傳 `nil` |

## Config 與 PresetConfig

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

欄位與預設值見 [設定](/zh/configuration) 與 [Preset](/zh/configuration-presets)。

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

數值越小越先執行。見 [優先權等級](/zh/core-concepts-priority) 與 [逾時推算](/zh/core-concepts-timeout)。

其他參考頁：[Enqueue 選項](/zh/api-reference-options)、[錯誤](/zh/api-reference-errors)、[日誌事件](/zh/api-reference-events)、[已移除 API](/zh/removed-api)。
