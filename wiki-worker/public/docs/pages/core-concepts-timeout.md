# Timeout Derivation

This page explains how go-queue computes each task's timeout from the base timeout, the preset's priority, and the 15–120 second clamp.

## Calculation Steps

The timeout is fixed at `Enqueue` time (`core/priority.go:17`); later promotion does not change it.

1. Base: `PresetConfig.Timeout` when `> 0`, otherwise `Config.Timeout`
2. Apply the multiplier for the preset's priority
3. Clamp to 15–120 seconds
4. `WithTimeout(d)` replaces the result with `d` and skips the clamp; a `d` of zero or less gives the task an already-expired ctx

## Multipliers

| Priority | Multiplier | With `Config.Timeout = 30s` |
|---|---|---|
| `PriorityImmediate` | ÷ 4 | 15s (7.5s raised to the floor) |
| `PriorityHigh` | ÷ 2 | 15s |
| `PriorityRetry` | ÷ 2 | 15s |
| `PriorityNormal` | × 1 | 30s |
| `PriorityLow` | × 2 | 60s |

The multiplier comes from the preset's priority at `Enqueue`. A retried task keeps its original timeout even though it re-enters at `PriorityRetry`.

When a task exceeds its timeout, see [Execution and Retries](/core-concepts-execution) for what happens to the task and its goroutine.
