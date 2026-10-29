# Preset

本頁說明 `PresetConfig`、Preset 如何決定任務的優先權與基準逾時，以及未定義的 Preset 名稱會解析成什麼。

## PresetConfig

```go
type PresetConfig struct {
	Priority Priority
	Timeout  time.Duration
}
```

| 欄位 | 零值 | 說明 |
|---|---|---|
| `Priority` | `PriorityImmediate` | 使用此 Preset 的任務初始優先權 |
| `Timeout` | `0` | 大於 0 時取代 `Config.Timeout`，作為此 Preset 的基準逾時 |

map 中不存在的 Preset 名稱（含 `""`）會解析為零值 `PresetConfig{}`，即 `PriorityImmediate` 加上全域基準逾時。

## 設定範例

```go
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
```

| Preset | 優先權 | 實際逾時 | 升級 |
|---|---|---|---|
| `payment` | High | `30s ÷ 2 = 15s` | 無 |
| `email` | Normal | `30s` | 等待 60 秒後升為 High |
| `report` | Low | `min(60s × 2, 120s) = 120s` | 等待 30 秒後升為 Normal |
| `""` 或未定義 | Immediate | `max(30s ÷ 4, 15s) = 15s` | 無 |

升級門檻只讀 `Config.Timeout`，不讀 Preset 的 `Timeout`。各欄背後的規則：[優先權等級](/zh/core-concepts-priority)、[逾時推算](/zh/core-concepts-timeout)、[防飢餓自動升級](/zh/core-concepts-promotion)。
