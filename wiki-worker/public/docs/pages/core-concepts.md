# Core Concepts

This page maps each scheduling question to the page that answers it explains where queued tasks live, and places go-queue among related Go projects; every rule below lives in `github.com/pardnchiu/go-queue/core`.

## Where to Look

| Question | Page |
|---|---|
| Which task runs first, and where does its priority come from? | [Priority Levels](/core-concepts-priority) |
| Can a low-priority task wait forever? | [Anti-Starvation Promotion](/core-concepts-promotion) |
| How long may a task run? | [Timeout Derivation](/core-concepts-timeout) |
| What happens on panic, timeout, or error, and how do retries work? | [Execution and Retries](/core-concepts-execution) |
| When can I call `Enqueue`, `Start`, and `Shutdown`? | [Queue Lifecycle](/core-concepts-lifecycle) |

## Task Path at a Glance

1. `Enqueue` resolves the preset's priority and timeout, then pushes the task into one bounded min-heap
2. An idle worker calls `Pop`; each `Pop` first promotes tasks that have waited too long
3. The worker runs the task under a timeout ctx with panic recovery
4. A failed task with retries left re-enters the heap at `PriorityRetry`; everything else ends in a `slog` event

## Where Tasks Live

go-queue is an in-process priority task queue: the bounded worker pool and the heap live inside your Go program, and the module imports only the standard library (`go.mod` has no `require`). Queued tasks are held in memory and are not persisted, so tasks still in the heap when the process exits never run. The table below places go-queue next to related Go projects.

## Related Go Projects

Each project's description comes from its own GitHub repository.

| Project | What It Is | How go-queue Relates |
|---|---|---|
| `hibiken/asynq` | Distributed task queue backed by Redis | go-queue keeps tasks in process memory and needs no external service |
| `riverqueue/river` | Background jobs stored in Postgres or SQLite | go-queue keeps tasks in process memory and needs no database |
| `RichardKnop/machinery` | Asynchronous task queue based on distributed message passing through brokers such as Redis, AMQP, or AWS SQS | go-queue runs inside one process with no broker |
| `panjf2000/ants`, `alitto/pond` | In-process goroutine pools | Same deployment shape; go-queue schedules from a priority heap with five priority levels, anti-starvation promotion, a per-task timeout derived from the preset, and optional retries |
| `susamn/rio` | Job scheduler based on a priority queue, with timeout and retry | Same deployment shape; go-queue picks priority and timeout per named preset (`Config.Preset`) and reports every outcome as a fixed `slog` event |
| `oleiade/lane` | Generic priority queue, queue, stack, and deque data structures | lane provides the data structure; go-queue is a task runner whose workers, timeouts, and retries sit on its own `container/heap` min-heap |
