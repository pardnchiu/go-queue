# Update Log

> Generated: 2025-12-31

## Summary

將優先級從字串型別重構為整數常數型別，優化 Retry 選項介面，新增 taskHeap 記憶體管理機制，並移除任務洩漏檢測邏輯。

## Changes

### REFACTOR
- **[priority.go]**: 將 priority 從字串型別改為整數常數（`PriorityImmediate`, `PriorityHigh`, `PriorityRetry`, `PriorityNormal`, `PriorityLow`），移除字串解析邏輯
- **[new.go]**: 將 `PresetConfig.Priority` 從 `string` 改為 `priority` 型別，`PresetConfig.Timeout` 從 `int` 改為 `time.Duration`
- **[option.go]**: 將 `WithRetry` 參數從 `*int` 改為變長參數 `...int`，支援無參數調用

### UPDATE
- **[priority.go]**: 改進 timeout 計算邏輯，直接使用 `time.Duration` 運算，並在函式內部進行邊界限制（15-120 秒）
- **[pending.go]**: 更新優先級升級邏輯，使用新的優先級常數替代字串

### PERF
- **[task.go]**: 新增 taskHeap cap 縮減機制，當使用量低於容量 1/4 時自動縮減至一半（最小保留 16）

### REMOVE
- **[new.go]**: 移除任務洩漏檢測邏輯（leakTimeout）
- **[priority.go]**: 移除 `getPresetPriority` 函式
- **[task.go]**: 在 `Pop()` 中設置 `old[n-1] = nil` 避免記憶體洩漏

### TEST
- **[main_test.go]**: 更新測試用例以使用新的優先級常數，修改 `WithRetry` 調用方式

## Files Changed

| File | Status | Tag |
|------|--------|-----|
| `priority.go` | Modified | REFACTOR |
| `new.go` | Modified | REFACTOR, UPDATE, REMOVE |
| `option.go` | Modified | REFACTOR |
| `pending.go` | Modified | UPDATE |
| `task.go` | Modified | PERF |
| `main_test.go` | Modified | TEST |
