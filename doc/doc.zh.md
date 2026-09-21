# go-queue - 技術文件

> 返回 [README](./README.zh.md)

## 前置需求

- Go 1.23 或以上
- 無外部服務相依（僅使用標準函式庫）

## 安裝

### 使用 go get

```bash
go get github.com/pardnchiu/go-queue
```

```go
import "github.com/pardnchiu/go-queue/core"
```

> `core` 子套件路徑自 v1.1.3 之後的版本起生效；在新版標籤發布前，請改用 `go get github.com/pardnchiu/go-queue@develop`。

### 從原始碼

```bash
git clone https://github.com/pardnchiu/go-queue.git
cd go-queue
go test -race ./...
```

## 使用方式

### 基礎

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

	// 等待佇列排空，最多 10 秒
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := q.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
```

### 進階：Preset、重試與 Callback

```go
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/pardnchiu/go-queue/core"
)

func main() {
	// 開啟 Debug 等級以觀察 task.promoted / task.timeout_triggered 事件
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	q := core.New(&core.Config{
		Workers: 8,
		Size:    1024,
		Timeout: 30 * time.Second,
		Preset: map[string]core.PresetConfig{
			"payment": {Priority: core.PriorityHigh},
			"email":   {Priority: core.PriorityNormal},
			"report":  {Priority: core.PriorityLow, Timeout: 60 * time.Second},
		},
	})

	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	if err := q.Start(ctx); err != nil {
		log.Fatal(err)
	}

	// 失敗時最多重試 5 次，成功後觸發 Callback
	_, err := q.Enqueue(ctx, "payment", func(ctx context.Context) error {
		return charge(ctx)
	},
		core.WithTaskID("order-1001"),
		core.WithRetry(5),
		core.WithCallback(func(id string) {
			slog.Info("charged", "id", id)
		}),
	)
	if err != nil {
		log.Printf("enqueue payment: %v", err)
	}

	// WithTimeout 直接覆寫 Preset 推算出的逾時（不受 15–120 秒限制）
	_, err = q.Enqueue(ctx, "report", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return nil
		}
	}, core.WithTimeout(5*time.Minute))
	if err != nil {
		log.Printf("enqueue report: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := q.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

var errDeclined = errors.New("card declined")

func charge(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return errDeclined
}
```

### 行為須知

| 情境 | 行為 |
|------|------|
| `Start` 前呼叫 `Enqueue` | 允許；任務暫存於堆中，`Start` 後依優先權執行 |
| 任務未監聽 `ctx.Done()` | 逾時後 Worker 回報錯誤並繼續處理下一個任務，但該任務的 goroutine 會持續執行直到自行返回 |
| `Shutdown` 期間任務失敗並觸發重試 | 狀態已為 Closed，重新入列失敗，記錄 `task.retry_failed` 並放棄該任務 |
| 取消傳給 `Start` 的 ctx | 所有任務的 ctx 隨之取消；Worker 不會停止，後續任務會立即以 `context canceled` 失敗 |
| `Shutdown` 超過期限 | 取消所有任務 ctx 並回傳錯誤；剩餘任務以已取消的 ctx 執行 |
| Callback | 僅在任務成功時於獨立 goroutine 呼叫；失敗、逾時、耗盡重試皆不觸發 |

## API 參考

### 建構與生命週期

| 函式 | 簽章 | 說明 |
|------|------|------|
| `New` | `func New(config *Config) *Queue` | 建立佇列；`config` 可為 `nil`，零值欄位套用預設值 |
| `Start` | `func (q *Queue) Start(ctx context.Context) error` | 啟動 `Workers` 個 Worker；`ctx` 為所有任務 ctx 的父 ctx |
| `Enqueue` | `func (q *Queue) Enqueue(ctx context.Context, presetName string, action func(ctx context.Context) error, options ...EnqueueOption) (string, error)` | 將任務推入佇列，回傳任務 ID；`ctx` 僅用於入列前檢查，不傳入任務 |
| `Shutdown` | `func (q *Queue) Shutdown(ctx context.Context) error` | 拒絕新任務、等待 Worker 排空佇列；`ctx` 到期則提前返回錯誤。重複呼叫回傳 `nil` |

### Config

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

| 欄位 | 預設值 | 說明 |
|------|--------|------|
| `Workers` | `runtime.NumCPU() * 2` | 併發 Worker 數 |
| `Size` | `Workers * 64` | 堆容量上限（含重試任務），超過時 `Enqueue` 回傳錯誤 |
| `Timeout` | `30 * time.Second` | 全域基準逾時；也決定自動升級門檻 |
| `Preset` | 空 map | Preset 名稱 → `PresetConfig` |
| `PresetConfig.Priority` | `PriorityImmediate`（零值） | 未設定或 Preset 名稱不存在時為 `PriorityImmediate`，包含 `""` |
| `PresetConfig.Timeout` | `0` | `> 0` 時取代 `Config.Timeout` 作為該 Preset 的基準逾時 |

### Priority

數值越小越優先；同優先權依入列（或重新入列）時間排序。

| 常數 | 值 | 逾時倍率 |
|------|----|---------|
| `PriorityImmediate` | `0` | 基準 ÷ 4 |
| `PriorityHigh` | `1` | 基準 ÷ 2 |
| `PriorityRetry` | `2` | 基準 ÷ 2 |
| `PriorityNormal` | `3` | 基準 × 1 |
| `PriorityLow` | `4` | 基準 × 2 |

**逾時計算：** `clamp(基準 × 倍率, 15s, 120s)`，基準為 `PresetConfig.Timeout`（若 `> 0`）否則 `Config.Timeout`。倍率依 Preset 設定的優先權決定，入列後的升級不會改變逾時。`WithTimeout` 直接覆寫結果，不套用 clamp。

**自動升級：** 每次 Worker 取任務時檢查整個堆。

| 由 | 至 | 等待門檻 |
|----|----|---------|
| `PriorityLow` | `PriorityNormal` | `clamp(Config.Timeout, 30s, 120s)` |
| `PriorityNormal` | `PriorityHigh` | `clamp(Config.Timeout × 2, 30s, 120s)` |

### EnqueueOption

| 選項 | 簽章 | 說明 |
|------|------|------|
| `WithTaskID` | `func WithTaskID(id string) EnqueueOption` | 自訂任務 ID；未指定時產生 UUID v4 |
| `WithTimeout` | `func WithTimeout(d time.Duration) EnqueueOption` | 覆寫該任務的逾時 |
| `WithCallback` | `func WithCallback(fn func(id string)) EnqueueOption` | 任務成功後以任務 ID 呼叫 |
| `WithRetry` | `func WithRetry(retryMax ...int) EnqueueOption` | 啟用重試；未帶參數時最多重試 3 次。總執行次數 = `retryMax + 1` |

重試立即以 `PriorityRetry` 重新入列，不含退避（Backoff）延遲。

### 錯誤

| 來源 | 錯誤訊息 |
|------|---------|
| `Start` | `queue already started`、`queue already closed` |
| `Enqueue` | `ctx.Err()`、`enqueue failed: staging queue is full`、`enqueue failed: staging queue is closed` |
| `Shutdown` | `shutdown timeout: N tasks remaining` |
| 任務執行 | `task timeout after <d>`、`panic: <value>` |

### slog 事件

| 事件 | 等級 | 觸發時機 |
|------|------|---------|
| `task.completed` | Info | 任務成功 |
| `task.retrying` | Warn | 任務失敗且仍有重試次數 |
| `task.failed` | Error | 未啟用重試的任務失敗 |
| `task.exhausted` | Error | 重試次數耗盡 |
| `task.retry_failed` | Error | 重新入列失敗（佇列已滿或已關閉） |
| `task.promoted` | Debug | 任務自動升級 |
| `task.timeout_triggered` | Debug | 任務逾時 |

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
