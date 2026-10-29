# Anti-Starvation Promotion

This page explains how go-queue lifts long-waiting Low and Normal tasks so a steady stream of urgent work cannot starve them.

## Promotion Rules

Each time a worker calls `Pop`, `promoteLocked` scans the whole heap and lifts long-waiting tasks one level (`core/pending.go:111`).

| From | To | Wait Threshold |
|---|---|---|
| `PriorityLow` | `PriorityNormal` | `clamp(Config.Timeout, 30s, 120s)` |
| `PriorityNormal` | `PriorityHigh` | `clamp(Config.Timeout × 2, 30s, 120s)` |

With the default `Config.Timeout = 30s`, a Low task becomes Normal after 30s, and a Normal task becomes High after 60s.

## Boundaries

- Wait time counts from `startAt`, which resets on retry
- Immediate, High, and Retry never promote
- A promoted Low task keeps its original `startAt`, so once it also passes the Normal threshold it moves on to High
- Thresholds read only `Config.Timeout`, never a preset's `Timeout`
- Promotion only runs on `Pop`, so no promotion checks happen while every worker is busy
- Promotion changes the order only; the timeout fixed at `Enqueue` stays the same
- Each promotion emits a Debug-level `task.promoted` event, see [Log Events](/api-reference-events)
