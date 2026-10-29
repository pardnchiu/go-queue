# 設定

本頁說明佇列層級 `Config` 的每個欄位、預設值與零值解析方式。go-queue 不讀環境變數或設定檔，所有設定都以 struct 傳入 `core.New`。

## Config

```go
type Config struct {
	Workers int
	Size    int
	Timeout time.Duration
	Preset  map[string]PresetConfig
}
```

| 欄位 | 零值時的預設 | 說明 |
|---|---|---|
| `Workers` | `runtime.NumCPU() * 2` | 並行 Worker 數 |
| `Size` | `Workers * 64`（以預設 Workers 計算） | 堆容量，含重試；滿了之後 `Enqueue` 回傳錯誤 |
| `Timeout` | `30 * time.Second` | 全域基準逾時；也決定升級門檻 |
| `Preset` | 空 map | Preset 名稱 → `PresetConfig`；`New` 會複製一份，之後修改原 map 不影響佇列 |

## 零值解析

| 輸入 | 結果 |
|---|---|
| `New(nil)` | 全部採用預設 |
| 只設 `Workers: 1` | `Size` 仍為 `NumCPU * 2 * 64`：`Size` 預設由 `runtime.NumCPU() * 2` 推得，而非你傳入的 `Workers`（`core/new.go:44`） |
| 負值 | 不會被拒絕；非零值一律照用 |

具名 Preset 見 [Preset](/zh/configuration-presets)；單一任務的調整見 [Enqueue 選項](/zh/api-reference-options)。
