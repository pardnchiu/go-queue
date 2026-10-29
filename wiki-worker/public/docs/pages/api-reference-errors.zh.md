# 錯誤

本頁列出 go-queue 回傳或記錄的所有錯誤訊息、觸發條件，以及比對方式。

## 錯誤訊息

| 來源 | 訊息 | 觸發條件 |
|---|---|---|
| `Start` | `queue already started` | 已在 Running |
| `Start` | `queue already closed` | 已 Closed |
| `Enqueue` | 原樣回傳 `ctx.Err()` | 傳入的 ctx 已取消或到期 |
| `Enqueue` | `enqueue failed: staging queue is full` | 堆長度達到 `Config.Size` |
| `Enqueue` | `enqueue failed: staging queue is closed` | 已呼叫 `Shutdown` |
| `Shutdown` | `shutdown timeout: N tasks remaining` | 堆排空前 `ctx` 到期 |
| 任務（僅記錄） | `task timeout after <d>` | 任務超過逾時 |
| 任務（僅記錄） | `panic: <value>` | 任務函式 panic |
| 任務（僅記錄） | `context canceled` | `Start` 的 ctx 被取消，或 `Shutdown` 期限到時任務仍在執行或排隊中 |

## 比對錯誤

`Enqueue` 的 ctx 錯誤不經包裝直接回傳，因此 `errors.Is(err, context.Canceled)` 與 `errors.Is(err, context.DeadlineExceeded)` 可用。佇列已滿／已關閉的錯誤沒有公開 sentinel，只能比對訊息：

```go
if _, err := q.Enqueue(ctx, "email", sendEmail); err != nil {
	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		// 呼叫端放棄
	case strings.Contains(err.Error(), "staging queue is full"):
		// 退避後再試
	case strings.Contains(err.Error(), "staging queue is closed"):
		// 佇列正在關閉
	}
}
```

任務端的錯誤不會回到呼叫端，只會出現在 [日誌事件](/zh/api-reference-events)。
