# 優先權等級

本頁說明五個優先權等級、同級時的排序方式，以及任務的優先權從哪裡來。

## 五級優先權

所有任務放在同一個最小堆中。數值越小越先執行；同級時依 `startAt`（入列或重新入列時間）先進先出（`core/task.go:36`）。

| 常數 | 數值 | 用途 |
|---|---|---|
| `PriorityImmediate` | `0` | 最先執行；也是 `Priority` 的零值 |
| `PriorityHigh` | `1` | 高優先；Normal 升級後的落點 |
| `PriorityRetry` | `2` | 重試專用，排在一般流量之前 |
| `PriorityNormal` | `3` | 一般任務；Low 升級後的落點 |
| `PriorityLow` | `4` | 背景任務 |

## 任務如何取得優先權

任務的優先權取自 `Enqueue` 傳入的 `presetName` 所對應 `PresetConfig.Priority`（`core/new.go:252`）；沒有逐任務指定優先權的選項。**Preset 名稱不存在（含 `""`）或 Preset 未設定 `Priority` 時，任務會拿到零值 `PriorityImmediate`。** 一般優先權的工作請定義 Preset 並明確設定 `PriorityNormal`。

| 事件 | 之後的優先權 |
|---|---|
| `Enqueue` | Preset 的 `Priority` |
| 等待超過升級門檻 | 升一級，見 [防飢餓自動升級](/zh/core-concepts-promotion) |
| 失敗且仍有重試額度 | `PriorityRetry`，見 [執行與重試](/zh/core-concepts-execution) |

Preset 的定義方式見 [Preset](/zh/configuration-presets)。
