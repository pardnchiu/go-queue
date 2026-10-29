# 已移除 API

本頁記錄 go-queue 移除過的每個符號、移除的版本，以及替代做法。

## 根套件搬移至 `core`

所有公開符號已從模組根 `github.com/pardnchiu/go-queue` 搬到子套件 `github.com/pardnchiu/go-queue/core`（commit `bafc412`）。簽章與行為不變，只有 import 路徑不同。舊根套件名為 `goQueue`，因此 v1.1.3 的程式碼通常以 `queue` 等別名匯入。

| 符號 | 套件 | 狀態 | 最後存在版本 | 替代 |
|---|---|---|---|---|
| `New` | `github.com/pardnchiu/go-queue` | 移除於 v1.1.3 之後的下一個版本（未發版） | v1.1.3 | `core.New` |
| `Queue`（`Start`、`Enqueue`、`Shutdown`） | `github.com/pardnchiu/go-queue` | 移除於 v1.1.3 之後的下一個版本（未發版） | v1.1.3 | `core.Queue` |
| `Config` | `github.com/pardnchiu/go-queue` | 移除於 v1.1.3 之後的下一個版本（未發版） | v1.1.3 | `core.Config` |
| `PresetConfig` | `github.com/pardnchiu/go-queue` | 移除於 v1.1.3 之後的下一個版本（未發版） | v1.1.3 | `core.PresetConfig` |
| `Priority` 及其常數 | `github.com/pardnchiu/go-queue` | 移除於 v1.1.3 之後的下一個版本（未發版） | v1.1.3 | `core.Priority`、`core.PriorityImmediate` ... `core.PriorityLow` |
| `EnqueueOption` | `github.com/pardnchiu/go-queue` | 移除於 v1.1.3 之後的下一個版本（未發版） | v1.1.3 | `core.EnqueueOption` |
| `WithTaskID`、`WithTimeout`、`WithCallback`、`WithRetry` | `github.com/pardnchiu/go-queue` | 移除於 v1.1.3 之後的下一個版本（未發版） | v1.1.3 | `core.WithTaskID`、`core.WithTimeout`、`core.WithCallback`、`core.WithRetry` |
| 未公開型別上的方法（`pending`：`Push`、`Pop`、`Len`、`State`、`Close`；`taskHeap`：`Len`、`Less`、`Swap`、`Push`、`Pop`） | `github.com/pardnchiu/go-queue` | 移除於 v1.1.3 之後的下一個版本（未發版）；呼叫端從未能存取 | v1.1.3 | 無公開替代；原樣搬入 `core` |

## 遷移

```diff
-import queue "github.com/pardnchiu/go-queue"
+import "github.com/pardnchiu/go-queue/core"

-q := queue.New(&queue.Config{Workers: 4})
+q := core.New(&core.Config{Workers: 4})
```
