# Priority Levels

This page explains the five priority levels, how ties break, and where a task's priority comes from.

## Five Priority Levels

Every task lives in one min-heap. Lower values run first; ties break FIFO on `startAt`, the enqueue or re-enqueue time (`core/task.go:36`).

| Constant | Value | Purpose |
|---|---|---|
| `PriorityImmediate` | `0` | Runs first; also the zero value of `Priority` |
| `PriorityHigh` | `1` | High priority; where promoted Normal tasks end up |
| `PriorityRetry` | `2` | Reserved for retries, ahead of regular traffic |
| `PriorityNormal` | `3` | Regular tasks; where promoted Low tasks end up |
| `PriorityLow` | `4` | Background tasks |

## How a Task Gets Its Priority

A task's priority comes from the `PresetConfig.Priority` of the `presetName` passed to `Enqueue` (`core/new.go:252`); there is no per-task priority option. **When the preset name is unknown (including `""`) or the preset leaves `Priority` unset, the task gets the zero value `PriorityImmediate`.** For regular-priority work, define a preset and set `PriorityNormal` explicitly.

| Event | Priority Afterwards |
|---|---|
| `Enqueue` | The preset's `Priority` |
| Waited past the promotion threshold | One level up, see [Anti-Starvation Promotion](/core-concepts-promotion) |
| Failed with retries left | `PriorityRetry`, see [Execution and Retries](/core-concepts-execution) |

Presets are defined in [Presets](/configuration-presets).
