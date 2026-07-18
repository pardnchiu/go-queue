# go-queue - 技術文件

> 返回 [README](./README.zh.md)

## 前置需求

- Go 1.23 或更高版本

## 安裝

### 使用 go get

```bash
go get github.com/pardnchiu/go-queue
```

### 從原始碼

```bash
git clone https://github.com/pardnchiu/go-queue.git
cd go-queue
go test ./...
```

## 使用方式

### 基礎用法

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

### 優先級預設

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

### 入隊選項

```go
id, err := q.Enqueue(ctx, "urgent", func(ctx context.Context) error {
	// 業務邏輯
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

### 優雅關閉

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := q.Shutdown(shutdownCtx); err != nil {
	// 逾時：仍有任務未完成
	log.Printf("shutdown: %v", err)
}
```

## API 參考

### New

```go
func New(config *Config) *Queue
```

建立佇列。`config` 可為 `nil`，此時使用預設值。

| 欄位 | 型別 | 預設 | 說明 |
|------|------|------|------|
| `Workers` | `int` | `runtime.NumCPU() * 2` | Worker 數量 |
| `Size` | `int` | `Workers * 64` | 待處理佇列容量 |
| `Timeout` | `time.Duration` | `30 * time.Second` | 基準逾時 |
| `Preset` | `map[string]PresetConfig` | empty | 具名優先級／逾時設定 |

### PresetConfig

```go
type PresetConfig struct {
	Priority Priority
	Timeout  time.Duration
}
```

| 欄位 | 說明 |
|------|------|
| `Priority` | 優先級；零值為 `PriorityImmediate`（iota 0） |
| `Timeout` | `0` 時依 Priority 由基準 `Timeout` 推算 |

### Priority

| 常數 | 值 | 說明 |
|------|----|------|
| `PriorityImmediate` | 0 | 最高，立刻執行 |
| `PriorityHigh` | 1 | 高優先 |
| `PriorityRetry` | 2 | 重試任務 |
| `PriorityNormal` | 3 | 一般 |
| `PriorityLow` | 4 | 最低；逾時後可晉升 |

### Start

```go
func (q *Queue) Start(ctx context.Context) error
```

啟動 worker 池。僅能從 `Created` 轉為 `Running`；重複呼叫或關閉後呼叫會回傳錯誤。

### Enqueue

```go
func (q *Queue) Enqueue(ctx context.Context, presetName string, action func(ctx context.Context) error, options ...EnqueueOption) (string, error)
```

將任務入隊並回傳 task ID。佇列已關閉、已滿，或 `ctx` 已取消時回傳錯誤。

### EnqueueOption

| 選項 | 簽章 | 說明 |
|------|------|------|
| `WithTaskID` | `func WithTaskID(id string) EnqueueOption` | 自訂 task ID；省略則自動產生 UUID |
| `WithTimeout` | `func WithTimeout(d time.Duration) EnqueueOption` | 覆寫本次任務逾時 |
| `WithCallback` | `func WithCallback(fn func(id string)) EnqueueOption` | 成功完成後非同步回呼 |
| `WithRetry` | `func WithRetry(retryMax ...int) EnqueueOption` | 啟用重試；省略參數時預設最多 3 次 |

### Shutdown

```go
func (q *Queue) Shutdown(ctx context.Context) error
```

關閉佇列、排空待處理任務並等待 worker 結束。`ctx` 逾時時回傳剩餘任務數錯誤。可重複呼叫（冪等）。

### 逾時推算規則

基準為 `Config.Timeout`（或 preset 覆寫），再依優先級調整，並限制在 15s–120s：

| Priority | 計算 |
|----------|------|
| Immediate | `timeout / 4` |
| High / Retry | `timeout / 2` |
| Normal | `timeout` |
| Low | `timeout * 2` |

### 優先級晉升

| 來源 | 等待時間 | 目標 |
|------|----------|------|
| Low | `clamp(Timeout, 30s, 120s)` | Normal |
| Normal | `clamp(Timeout*2, 30s, 120s)` | High |

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
