# go-queue - 技術文件

> 返回 [README](./README.zh.md)

## 前置需求

- Go 1.23 或更高版本

## 安裝

### 使用 go get

```bash
go get github.com/pardnchiu/go-queue
```

### 從原始碼建置

```bash
git clone https://github.com/pardnchiu/go-queue.git
cd go-queue
go test ./...
```

## 使用方式

### 基礎用法

建立佇列、啟動 Worker 並提交任務：

```go
package main

import (
	"context"
	"fmt"
	"log"

	goQueue "github.com/pardnchiu/go-queue"
)

func main() {
	q := goQueue.New(&goQueue.Config{
		Workers: 4,
	})

	ctx := context.Background()
	if err := q.Start(ctx); err != nil {
		log.Fatal(err)
	}

	id, err := q.Enqueue(ctx, "", func(ctx context.Context) error {
		fmt.Println("task executed")
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("task id:", id)

	if err := q.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
```

### 使用 Preset 與優先級

透過 Preset 預定義不同任務類型的優先級與 Timeout：

```go
q := goQueue.New(&goQueue.Config{
	Workers: 4,
	Timeout: 30 * time.Second,
	Preset: map[string]goQueue.PresetConfig{
		"critical": {Priority: goQueue.PriorityImmediate},
		"batch":    {Priority: goQueue.PriorityLow, Timeout: 60 * time.Second},
	},
})

ctx := context.Background()
q.Start(ctx)

// 以 Immediate 優先級入列
q.Enqueue(ctx, "critical", func(ctx context.Context) error {
	// 高優先任務
	return nil
})

// 以 Low 優先級入列，Timeout 60 秒
q.Enqueue(ctx, "batch", func(ctx context.Context) error {
	// 低優先批次任務
	return nil
})

q.Shutdown(ctx)
```

### 使用 Retry 與 Callback

為任務配置重試機制與完成回呼：

```go
q := goQueue.New(&goQueue.Config{Workers: 2})
ctx := context.Background()
q.Start(ctx)

id, err := q.Enqueue(ctx, "", func(ctx context.Context) error {
	// 可能失敗的操作
	return doSomething()
},
	goQueue.WithRetry(3),                           // 最多重試 3 次
	goQueue.WithTaskID("my-task-001"),               // 自訂任務 ID
	goQueue.WithTimeout(10*time.Second),             // 任務級 Timeout
	goQueue.WithCallback(func(id string) {           // 完成回呼
		fmt.Printf("task %s completed\n", id)
	}),
)
if err != nil {
	log.Fatal(err)
}

q.Shutdown(ctx)
```

### Graceful Shutdown 與 Timeout

限制 Shutdown 等待時間，避免無限阻塞：

```go
q := goQueue.New(&goQueue.Config{Workers: 4})
ctx := context.Background()
q.Start(ctx)

// ... 提交任務 ...

shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := q.Shutdown(shutdownCtx); err != nil {
	fmt.Printf("shutdown timeout: %v\n", err)
}
```

## API 參考

### New

```go
func New(config *Config) *Queue
```

建立新的佇列實例。`config` 為 `nil` 時使用預設值。

### Config

```go
type Config struct {
	Workers int                     // Worker 數量，預設 CPU * 2
	Size    int                     // 佇列容量上限，預設 Workers * 64
	Timeout time.Duration           // 全域任務 Timeout，預設 30 秒
	Preset  map[string]PresetConfig // 任務類型預設配置
}
```

### PresetConfig

```go
type PresetConfig struct {
	Priority Priority      // 優先等級，預設 PriorityNormal
	Timeout  time.Duration // 任務 Timeout，0 = 依 Priority 自動計算
}
```

### Priority

```go
type Priority int

const (
	PriorityImmediate Priority = iota // 最高優先
	PriorityHigh                       // 高優先
	PriorityRetry                      // 重試任務專用
	PriorityNormal                     // 一般（預設）
	PriorityLow                        // 低優先
)
```

Priority 影響 Timeout 計算：Immediate = base/4、High = base/2、Normal = base、Low = base*2，最終限制在 15~120 秒區間。

### Queue.Start

```go
func (q *Queue) Start(ctx context.Context) error
```

啟動 Worker Pool。重複呼叫或佇列已關閉時回傳錯誤。

### Queue.Enqueue

```go
func (q *Queue) Enqueue(ctx context.Context, presetName string, action func(ctx context.Context) error, options ...EnqueueOption) (string, error)
```

提交任務至佇列。`presetName` 對應 `Config.Preset` 中的鍵，空字串使用預設值。回傳任務 ID 或錯誤（Context 已取消、佇列已滿或已關閉）。

### Queue.Shutdown

```go
func (q *Queue) Shutdown(ctx context.Context) error
```

關閉佇列並等待所有 Worker 完成。Context 超時時回傳剩餘任務數量的錯誤。

### EnqueueOption

| 函式 | 簽章 | 說明 |
|------|------|------|
| `WithTaskID` | `func WithTaskID(id string) EnqueueOption` | 自訂任務 ID，預設自動產生 UUID v4 |
| `WithTimeout` | `func WithTimeout(d time.Duration) EnqueueOption` | 覆寫任務級 Timeout |
| `WithCallback` | `func WithCallback(fn func(id string)) EnqueueOption` | 任務成功完成後的回呼函式 |
| `WithRetry` | `func WithRetry(retryMax ...int) EnqueueOption` | 啟用重試，預設最多 3 次；可指定上限 |

***

©️ 2025 [邱敬幃 Pardn Chiu](https://linkedin.com/in/pardnchiu)
