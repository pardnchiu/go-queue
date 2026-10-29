# 逾時推算

本頁說明 go-queue 如何由基準逾時、Preset 優先權與 15–120 秒上下限算出每個任務的逾時。

## 計算步驟

逾時在 `Enqueue` 時就固定（`core/priority.go:17`），之後的升級不會改變它。

1. 基準：`PresetConfig.Timeout` 大於 0 時採用，否則用 `Config.Timeout`
2. 套用 Preset 優先權對應的倍率
3. 限制在 15–120 秒之間
4. `WithTimeout(d)` 直接以 `d` 取代結果，且不受上下限限制；`d` 小於等於 0 時任務拿到的是已到期的 ctx

## 倍率

| 優先權 | 倍率 | `Config.Timeout = 30s` 時 |
|---|---|---|
| `PriorityImmediate` | ÷ 4 | 15s（7.5s 拉到下限） |
| `PriorityHigh` | ÷ 2 | 15s |
| `PriorityRetry` | ÷ 2 | 15s |
| `PriorityNormal` | × 1 | 30s |
| `PriorityLow` | × 2 | 60s |

倍率取自 `Enqueue` 當下 Preset 的優先權。重試的任務雖以 `PriorityRetry` 重新入堆，仍沿用原本的逾時。

任務逾時後任務本身與其 goroutine 的去向，見 [執行與重試](/zh/core-concepts-execution)。
