# Removed API

This page records every symbol removed from go-queue, the version it left in, and what to use instead.

## Root Package Moved to `core`

Every exported symbol moved from the module root `github.com/pardnchiu/go-queue` to the sub-package `github.com/pardnchiu/go-queue/core` (commit `bafc412`). Signatures and behavior are unchanged; only the import path differs. The old root package was named `goQueue`, so v1.1.3 code usually imported it under an alias such as `queue`.

| Symbol | Package | Status | Last Present | Replacement |
|---|---|---|---|---|
| `New` | `github.com/pardnchiu/go-queue` | Removed in the first release after v1.1.3 (unreleased) | v1.1.3 | `core.New` |
| `Queue` (`Start`, `Enqueue`, `Shutdown`) | `github.com/pardnchiu/go-queue` | Removed in the first release after v1.1.3 (unreleased) | v1.1.3 | `core.Queue` |
| `Config` | `github.com/pardnchiu/go-queue` | Removed in the first release after v1.1.3 (unreleased) | v1.1.3 | `core.Config` |
| `PresetConfig` | `github.com/pardnchiu/go-queue` | Removed in the first release after v1.1.3 (unreleased) | v1.1.3 | `core.PresetConfig` |
| `Priority` and its constants | `github.com/pardnchiu/go-queue` | Removed in the first release after v1.1.3 (unreleased) | v1.1.3 | `core.Priority`, `core.PriorityImmediate` ... `core.PriorityLow` |
| `EnqueueOption` | `github.com/pardnchiu/go-queue` | Removed in the first release after v1.1.3 (unreleased) | v1.1.3 | `core.EnqueueOption` |
| `WithTaskID`, `WithTimeout`, `WithCallback`, `WithRetry` | `github.com/pardnchiu/go-queue` | Removed in the first release after v1.1.3 (unreleased) | v1.1.3 | `core.WithTaskID`, `core.WithTimeout`, `core.WithCallback`, `core.WithRetry` |
| Methods on unexported types (`pending`: `Push`, `Pop`, `Len`, `State`, `Close`; `taskHeap`: `Len`, `Less`, `Swap`, `Push`, `Pop`) | `github.com/pardnchiu/go-queue` | Removed in the first release after v1.1.3 (unreleased); never reachable by callers | v1.1.3 | No public replacement; moved unchanged into `core` |

## Migrating

```diff
-import queue "github.com/pardnchiu/go-queue"
+import "github.com/pardnchiu/go-queue/core"

-q := queue.New(&queue.Config{Workers: 4})
+q := core.New(&core.Config{Workers: 4})
```
