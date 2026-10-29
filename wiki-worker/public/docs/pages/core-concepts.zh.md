# 核心概念

本頁把每個排程問題對應到回答它的頁面，說明排隊中的任務存放在哪裡，並列出與相關 Go 專案的關係；以下規則全部位於 `github.com/pardnchiu/go-queue/core`。

## 該看哪一頁

| 問題 | 頁面 |
|---|---|
| 哪個任務先執行？優先權從哪來？ | [優先權等級](/zh/core-concepts-priority) |
| 低優先權任務會不會永遠等不到？ | [防飢餓自動升級](/zh/core-concepts-promotion) |
| 一個任務最多能跑多久？ | [逾時推算](/zh/core-concepts-timeout) |
| panic、逾時、回傳錯誤時會怎樣？重試怎麼運作？ | [執行與重試](/zh/core-concepts-execution) |
| 什麼時候可以呼叫 `Enqueue`、`Start`、`Shutdown`？ | [佇列生命週期](/zh/core-concepts-lifecycle) |

## 任務路徑速覽

1. `Enqueue` 依 Preset 決定優先權與逾時，再把任務放進單一有界最小堆（min-heap）
2. 閒置的 Worker 呼叫 `Pop`；每次 `Pop` 都會先升級等待過久的任務
3. Worker 在逾時 ctx 下執行任務，並攔截 panic
4. 失敗且仍有重試額度的任務以 `PriorityRetry` 重新入堆；其餘結果都以一筆 `slog` 事件收尾

## 任務存放在哪裡

go-queue 是程序內（in-process）的優先權任務佇列：有界 worker pool 與最小堆都在你的 Go 程式內，模組只引用標準函式庫（`go.mod` 沒有 `require`）。排隊中的任務只存在記憶體、不會持久化，程序結束時仍在堆裡的任務不會執行。下表列出 go-queue 與相關 Go 專案的關係。

## 相關 Go 專案

各專案的描述取自其 GitHub repository。

| 專案 | 定位 | 與 go-queue 的關係 |
|---|---|---|
| `hibiken/asynq` | 以 Redis 為後端的分散式任務佇列 | go-queue 把任務放在程序記憶體，不需要外部服務 |
| `riverqueue/river` | 存放在 Postgres 或 SQLite 的背景工作 | go-queue 把任務放在程序記憶體，不需要資料庫 |
| `RichardKnop/machinery` | 透過 Redis、AMQP、AWS SQS 等 broker 分散式傳遞訊息的非同步任務佇列 | go-queue 在單一程序內執行，不需要 broker |
| `panjf2000/ants`、`alitto/pond` | 程序內的 goroutine pool | 部署型態相同；go-queue 以優先權堆排程，提供五級優先權、防飢餓自動升級、依 Preset 推算的單一任務逾時與選用的重試 |
| `susamn/rio` | 以優先佇列為基礎、具逾時與重試的工作排程器 | 部署型態相同；go-queue 依具名 Preset（`Config.Preset`）決定優先權與逾時，每種結果都以固定的 `slog` 事件回報 |
| `oleiade/lane` | 泛型的優先佇列、佇列、堆疊與雙端佇列資料結構 | lane 提供資料結構；go-queue 是任務執行器，Worker、逾時與重試建立在自身的 `container/heap` 最小堆上 |
