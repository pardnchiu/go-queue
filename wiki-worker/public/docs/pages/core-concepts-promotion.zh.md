# 防飢餓自動升級

本頁說明 go-queue 如何提升等待過久的 Low 與 Normal 任務，避免持續湧入的緊急工作讓它們永遠排不到。

## 升級規則

每次 Worker 呼叫 `Pop`，`promoteLocked` 會掃描整個堆，把等待過久的任務升一級（`core/pending.go:111`）。

| 原等級 | 升至 | 等待門檻 |
|---|---|---|
| `PriorityLow` | `PriorityNormal` | `clamp(Config.Timeout, 30s, 120s)` |
| `PriorityNormal` | `PriorityHigh` | `clamp(Config.Timeout × 2, 30s, 120s)` |

使用預設 `Config.Timeout = 30s` 時，Low 任務等 30 秒升為 Normal，Normal 任務等 60 秒升為 High。

## 邊界

- 等待時間從 `startAt` 起算，重試時會重設
- Immediate、High、Retry 不會升級
- 升級後的 Low 任務沿用原本的 `startAt`，等待時間再超過 Normal 門檻時會接著升到 High
- 門檻只讀 `Config.Timeout`，不讀 Preset 的 `Timeout`
- 升級只在 `Pop` 時執行，所有 Worker 都在忙時不會有升級檢查
- 升級只改變順序；`Enqueue` 時決定的逾時不變
- 每次升級發出一筆 Debug 等級的 `task.promoted` 事件，見 [日誌事件](/zh/api-reference-events)
