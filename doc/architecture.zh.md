# go-queue - 架構

> 返回 [README](./README.zh.md)

## 概覽

```mermaid
graph TB
    App[應用程式] --> Enqueue[Enqueue]
    App --> Start[Start]
    App --> Shutdown[Shutdown]
    Enqueue --> Pending[待處理 Heap]
    Start --> Workers[Worker 池]
    Pending --> Workers
    Workers --> Execute[執行 / 逾時 / 重試]
    Shutdown --> Pending
    Shutdown --> Workers
```

## Module: Queue

對外 API 與生命週期入口，持有設定、pending 佇列與 worker 協調。

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
    App[呼叫端] --> New
    App --> Start
    App --> Enqueue
    App --> Shutdown
```

## Module: pending

以 mutex + cond 保護的優先佇列，支援晉升與關閉廣播。

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

`container/heap` 實作，依 priority 再依 startAt 排序，並在縮減時回收容量。

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

## 資料流

```mermaid
sequenceDiagram
    participant App as 應用程式
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
    alt 成功
        T-->>W: nil
        W->>App: callback(id) 非同步
    else 失敗且可重試
        W->>P: Push(PriorityRetry)
    else 失敗或逾時
        W-->>W: slog error
    end
    App->>Q: Shutdown(ctx)
    Q->>P: Close()
    Q->>W: wait WaitGroup
```

## 狀態機

```mermaid
stateDiagram-v2
    [*] --> Created
    Created --> Running: Start CAS
    Running --> Closed: Shutdown CAS
    Created --> Closed: Shutdown CAS
    Closed --> Closed: Shutdown 冪等
```

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
