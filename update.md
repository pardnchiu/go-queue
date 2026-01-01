# Update Log

> Generated: 2026-01-01

## Recommended Commit Message

refactor: 使用 atomic 狀態管理取代 mutex
refactor: replace mutex with atomic state management for thread-safe queue lifecycle

***

## Summary

將 Queue 狀態管理從 mutex + bool flag 重構為 atomic state machine，消除鎖競爭，確保狀態轉換原子性。同時統一 Timeout 型別為 `time.Duration`，簡化時間處理邏輯。

## Changes

### REFACTOR
- 新增 `queueState` 型別與狀態常數（`stateCreated`、`stateRunning`、`stateClosed`），實現狀態機
- 將 `Queue.closed` bool + `sync.RWMutex` 替換為 `atomic.Uint32` state
- 在 `New()` 中初始化 state 為 `stateCreated`
- 在 `Start()` 中使用 CAS 確保只能從 `stateCreated` 轉換到 `stateRunning`
- 在 `Shutdown()` 中使用 CAS loop 確保冪等性關閉（允許多次調用）
- 在 `setRetry()` 中使用 atomic load 檢查 `stateClosed` 狀態
- 將 `pending.closed` bool 替換為 `state *atomic.Uint32` 引用
- 修改 `newPending()` 簽名，接收 `queueState *atomic.Uint32` 參數
- 在 `Push()` 與 `Pop()` 中使用 atomic load 檢查狀態
- 新增 `State()` 方法返回當前 queue 狀態

### UPDATE
- 修改 `Start()` 返回型別為 `error`，支援狀態檢查錯誤回報
- `Config.Timeout` 型別從 `int64` 改為 `time.Duration`
- 預設 `Timeout` 值從 `30` 改為 `30 * time.Second`
- 移除 `getQueueTimeout()` 中所有 `time.Duration()` 型別轉換
- 移除 `getPromotion()` 中的 `time.Duration(c.Timeout) * time.Second` 轉換

### FIX
- 修正 `Enqueue()` 錯誤訊息格式化，使用 `%w` 包裹底層錯誤
- 修正 `Shutdown()` 在完成後呼叫 `cancel()` 清理 context

### REMOVE
- 移除所有 mutex 操作（`q.mu.Lock()`、`q.mu.Unlock()`）
- 移除 `Queue.mu` 與 `Queue.closed` 欄位
- 移除 `Close()` 中的 `p.closed = true` 賦值

***

## Summary

Refactor queue state management from mutex + bool flag to atomic state machine, eliminating lock contention and ensuring atomic state transitions. Unify `Timeout` type to `time.Duration` for simplified time handling.

## Changes

### REFACTOR
- Add `queueState` type and state constants (`stateCreated`, `stateRunning`, `stateClosed`)
- Replace `Queue.closed` bool + `sync.RWMutex` with `atomic.Uint32` state
- Initialize state to `stateCreated` in `New()`
- Use CAS in `Start()` to ensure transition from `stateCreated` to `stateRunning` only
- Use CAS loop in `Shutdown()` for idempotent closure (allow multiple calls)
- Use atomic load in `setRetry()` to check `stateClosed` state
- Replace `pending.closed` bool with `state *atomic.Uint32` reference
- Modify `newPending()` signature to accept `queueState *atomic.Uint32` parameter
- Use atomic load in `Push()` and `Pop()` for state checking
- Add `State()` method to return current queue state

### UPDATE
- Change `Start()` return type to `error` for state check error reporting
- Change `Config.Timeout` type from `int64` to `time.Duration`
- Change default `Timeout` from `30` to `30 * time.Second`
- Remove all `time.Duration()` type conversions in `getQueueTimeout()`
- Remove `time.Duration(c.Timeout) * time.Second` conversion in `getPromotion()`

### FIX
- Fix `Enqueue()` error message formatting, use `%w` to wrap underlying error
- Fix `Shutdown()` to call `cancel()` after completion for context cleanup

### REMOVE
- Remove all mutex operations (`q.mu.Lock()`, `q.mu.Unlock()`)
- Remove `Queue.mu` and `Queue.closed` fields
- Remove `p.closed = true` assignment in `Close()`

***

## Files Changed

| File | Status | Tag |
|------|--------|-----|
| `new.go` | Modified | REFACTOR, UPDATE, FIX, REMOVE |
| `pending.go` | Modified | REFACTOR, UPDATE, REMOVE |
| `priority.go` | Modified | UPDATE |