# 快速開始

本頁說明如何安裝 go-queue、啟動 Worker 池（Pool）、送出第一個任務，並在期限內關閉佇列。

## 前置需求

| 項目 | 需求 |
|---|---|
| Go | 1.23 以上（`go.mod` 宣告 `go 1.23`） |
| 外部相依 | 無，僅使用標準函式庫 |

## 安裝

```bash
go get github.com/pardnchiu/go-queue
```

```go
import "github.com/pardnchiu/go-queue/core"
```

公開 API 位於 `core` 子套件。此路徑自 v1.1.3 之後的版本起生效；新版標籤發布前請使用 `go get github.com/pardnchiu/go-queue@develop`。

## 第一個佇列

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

執行後每個任務會輸出一行 `task.completed` 的 `slog` 紀錄，內容包含任務 ID、Preset 名稱與耗時。

## 生命週期的三個呼叫

| 步驟 | 呼叫 | 重點 |
|---|---|---|
| 建立 | `core.New(cfg)` | `cfg` 可為 `nil`；零值欄位套用預設值 |
| 啟動 | `q.Start(ctx)` | 啟動 `Workers` 個 goroutine；`ctx` 取消時所有執行中任務的 ctx 一併取消 |
| 關閉 | `q.Shutdown(ctx)` | 先拒絕新任務，再等待 Worker 排空堆中剩餘任務；`ctx` 到期時回傳錯誤 |

`Start` 之前也能呼叫 `Enqueue`：任務會暫存在堆中，`Start` 後依優先權開始執行。

## 下一步

- 優先權等級與升級 → [優先權等級](/zh/core-concepts-priority)、[防飢餓自動升級](/zh/core-concepts-promotion)
- 任務最多能跑多久 → [逾時推算](/zh/core-concepts-timeout)
- `Config` 與 Preset 調整 → [設定](/zh/configuration)、[Preset](/zh/configuration-presets)
- 重試、Callback 與自訂 ID → [Enqueue 選項](/zh/api-reference-options)
- 完整函式簽章與錯誤訊息 → [API 參考](/zh/api-reference)、[錯誤](/zh/api-reference-errors)
