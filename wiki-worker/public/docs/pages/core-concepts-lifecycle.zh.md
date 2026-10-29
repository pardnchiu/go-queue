# 佇列生命週期

本頁說明佇列的三個狀態、各狀態允許哪些呼叫，以及 `Shutdown` 如何在期限內排空堆。

## Queue 狀態

狀態是一個 `atomic.Uint32`，所有轉換都透過 CompareAndSwap。

| 狀態 | 進入方式 | `Enqueue` | `Start` | `Shutdown` |
|---|---|---|---|---|
| Created | `New` | 允許（任務排隊等待） | 轉為 Running | 轉為 Closed |
| Running | `Start` | 允許 | `queue already started` | 轉為 Closed |
| Closed | `Shutdown` | `staging queue is closed` | `queue already closed` | 回傳 `nil` |

## 關閉

進入 Closed 後，`Shutdown` 讓 Worker 持續取任務直到堆清空，再取消根 ctx。若 `Shutdown` 的 ctx 先到期，會立即取消根 ctx 並回傳 `shutdown timeout: N tasks remaining`。

| 情境 | 結果 |
|---|---|
| 期限內排空 | 所有排隊中的任務都會執行；回傳 `nil` |
| 期限先到 | 執行中與仍在排隊的任務拿到已取消的 ctx；回傳 `shutdown timeout: N tasks remaining` |
| 重複呼叫 | 立即回傳 `nil` |
| 在 `Start` 前呼叫 | 轉為 Closed 並回傳 `nil`；排隊中的任務不會執行 |
