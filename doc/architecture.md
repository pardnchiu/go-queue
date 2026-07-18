# go-queue - Architecture

> Back to [README](../README.md)

## Overview

```mermaid
graph TB
    App[Application] --> Enqueue[Enqueue]
    App --> Start[Start]
    App --> Shutdown[Shutdown]
    Enqueue --> Pending[Pending Heap]
    Start --> Workers[Worker Pool]
    Pending --> Workers
    Workers --> Execute[Execute / Timeout / Retry]
    Shutdown --> Pending
    Shutdown --> Workers
```

## Module: Queue

Public API and lifecycle entry point. Owns config, pending queue, and worker coordination.

```mermaid
graph TB
    subgraph Queue
        New[New] --> Config[Config / Preset]
        New --> Pending[pending]
        Start[Start] --> Workers[worker goroutines]
        Enqueue[Enqueue] --> Options[EnqueueOption]
        Enqueue --> Pending
        Workers --> Execute[execute]
        Execute --> Retry[setRetry]
        Retry --> Pending
        Shutdown[Shutdown] --> Close[pending.Close]
        Shutdown --> Wait[WaitGroup]
    end
    App[Caller] --> New
    App --> Start
    App --> Enqueue
    App --> Shutdown
```

## Module: pending

Mutex + condition-variable protected priority queue with promotion and close broadcast.

```mermaid
graph TB
    subgraph pending
        Push[Push] --> Heap[taskHeap]
        Pop[Pop] --> Promote[promoteLocked]
        Promote --> Heap
        Pop --> Heap
        Close[Close] --> Broadcast[cond.Broadcast]
        Len[Len] --> Heap
    end
    Queue[Queue] --> Push
    Queue --> Pop
    Queue --> Close
```

## Module: taskHeap

`container/heap` ordered by priority then `startAt`, with capacity shrink on pop.

```mermaid
classDiagram
    class task {
        +string ID
        +string preset
        +Priority priority
        +func action
        +Duration timeout
        +func callback
        +Time startAt
        +bool retryOn
        +int retryMax
        +int retryTimes
    }
    class taskHeap {
        +taskHeaps tasks
        +int minCap
        +Len()
        +Less()
        +Swap()
        +Push()
        +Pop()
    }
    taskHeap --> task : holds
```

## Data Flow

```mermaid
sequenceDiagram
    participant App as Application
    participant Q as Queue
    participant P as pending
    participant W as Worker
    participant T as task action

    App->>Q: New(config)
    App->>Q: Start(ctx)
    Q->>W: spawn workers
    App->>Q: Enqueue(preset, action, opts)
    Q->>P: Push(task)
    P-->>W: Pop(task, promotions)
    W->>T: action(ctx with timeout)
    alt success
        T-->>W: nil
        W->>App: callback(id) async
    else failure with retry
        W->>P: Push(PriorityRetry)
    else failure or timeout
        W-->>W: slog error
    end
    App->>Q: Shutdown(ctx)
    Q->>P: Close()
    Q->>W: wait WaitGroup
```

## State Machine

```mermaid
stateDiagram-v2
    [*] --> Created
    Created --> Running: Start CAS
    Running --> Closed: Shutdown CAS
    Created --> Closed: Shutdown CAS
    Closed --> Closed: Shutdown idempotent
```

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
