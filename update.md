# Update Log

> Generated: 2026-01-01 20:00

## Summary

優化 taskHeap 記憶體管理，新增動態最小容量機制，完善 timeout 日誌，移除未使用的 leak detection。

## Changes

### REFACTOR
- **task.go**: 將 `taskHeap` 從 slice 型別重構為 struct，新增 `tasks` 與 `minCap` 欄位，支援動態最小容量
- **task.go**: 調整 `taskHeap` 的 `Len`、`Less`、`Swap`、`Push`、`Pop` 方法，適配新的 struct 結構
- **pending.go**: 修正 `newPending` 初始化，根據 `workers` 與 `size` 計算動態最小容量（`max(16, min(size/8, size/workers))`）

### PERF
- **task.go**: 優化 heap shrink 邏輯，將觸發條件由 `capacity > taskHeapMinCap && length < capacity/4` 改為 `capacity > minCap*4 && length < capacity/taskHeapMinCapRatio`
- **task.go**: 改進縮減容量計算，使用 `max(capacity/4, minCap)` 確保不低於最小容量
- **pending.go**: 優化 `promoteLocked` 存取效率，先取得 `tasks` 欄位避免重複存取

### UPDATE
- **new.go**: 修正 `newPending` 呼叫，新增 `workers` 參數
- **task.go**: 將 `taskHeapMinCap` 常數重新命名為 `taskHeapMinCapRatio`，數值保持為 8

### FIX
- **new.go**: 新增 timeout 觸發時的 debug 日誌（`task.timeout_triggered`），記錄 id、preset、timeout 資訊

### REMOVE
- **new.go**: 移除註解的 leak detection 程式碼（leakTimeout timer 與相關 select 邏輯）

## Files Changed

| File | Status | Tag |
|------|--------|-----|
| `task.go` | Modified | REFACTOR, PERF, UPDATE |
| `new.go` | Modified | UPDATE, FIX, REMOVE |
| `pending.go` | Modified | REFACTOR, PERF |