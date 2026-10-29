# Queue Lifecycle

This page explains the three queue states, which calls each state allows, and how `Shutdown` drains the heap within a deadline.

## Queue State

State is an `atomic.Uint32`, and every transition goes through CompareAndSwap.

| State | Entered By | `Enqueue` | `Start` | `Shutdown` |
|---|---|---|---|---|
| Created | `New` | Allowed (tasks wait) | Moves to Running | Moves to Closed |
| Running | `Start` | Allowed | `queue already started` | Moves to Closed |
| Closed | `Shutdown` | `staging queue is closed` | `queue already closed` | Returns `nil` |

## Shutdown

In the Closed state `Shutdown` lets workers keep popping until the heap is empty, then cancels the root ctx. If the `Shutdown` ctx expires first, it cancels the root ctx immediately and returns `shutdown timeout: N tasks remaining`.

| Situation | Result |
|---|---|
| Heap drains before the deadline | Every queued task runs; returns `nil` |
| Deadline expires first | Running and still-queued tasks get a cancelled ctx; returns `shutdown timeout: N tasks remaining` |
| Called again | Returns `nil` immediately |
| Called before `Start` | Moves to Closed and returns `nil`; queued tasks never run |
