package goQueue

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

type Queue struct {
	config  *Config
	pending *pending
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.RWMutex
	closed  bool
}

type Config struct {
	Workers int                     // default = CPU * 2
	Size    int                     // default = Workers * 64
	Timeout int64                   // default = 30
	Preset  map[string]PresetConfig // default = empty
}

type PresetConfig struct {
	Priority Priority      // nil = 用 PriorityNormal
	Timeout  time.Duration // 0 = 依 Priority 自動計算（秒）
}

func New(config *Config) *Queue {
	worker := runtime.NumCPU() * 2
	newConfig := &Config{
		Workers: worker,
		Size:    worker * 64,
		Timeout: 30,
		Preset:  make(map[string]PresetConfig),
	}

	if config != nil {
		if config.Workers != 0 {
			newConfig.Workers = config.Workers
		}
		if config.Size != 0 {
			newConfig.Size = config.Size
		}
		if config.Timeout != 0 {
			newConfig.Timeout = config.Timeout
		}
		if config.Preset != nil {
			newConfig.Preset = make(map[string]PresetConfig)
			for k, v := range config.Preset {
				newConfig.Preset[k] = v
			}
		}
	}

	return &Queue{
		config:  newConfig,
		pending: newPending(newConfig.Size, newConfig.getPromotion()),
	}
}

func (q *Queue) Start(ctx context.Context) {
	q.ctx, q.cancel = context.WithCancel(ctx)

	for i := 0; i < q.config.Workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

func (q *Queue) worker() {
	defer q.wg.Done()

	for {
		task, promotions, ok := q.pending.Pop()

		if !ok {
			return
		}

		for _, e := range promotions {
			slog.Debug("task.promoted", "id", e.taskID, "from", e.from, "to", e.to)
		}

		q.execute(task)
	}
}

func (q *Queue) execute(task *task) {
	ctx, cancel := context.WithTimeout(q.ctx, task.timeout)
	defer cancel()

	start := time.Now()

	type result struct {
		err error
	}
	done := make(chan result, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- result{err: fmt.Errorf("panic: %v", r)}
				return
			}
		}()
		err := task.action(ctx)
		done <- result{err: err}
	}()
	var err error
	select {
	case r := <-done:
		err = r.err
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			err = fmt.Errorf("task timeout after %s", task.timeout)
		} else {
			err = ctx.Err()
		}
		// leakTimeout := time.NewTimer(5 * time.Second)
		// select {
		// case <-done:
		// 	leakTimeout.Stop()
		// case <-leakTimeout.C:
		// 	slog.Warn("task.leaked", "id", task.ID, "preset", task.preset, "timeout", task.timeout)
		// }
	}

	elapsed := time.Since(start)

	if err != nil {
		if task.retryOn && task.retryTimes < task.retryMax {
			if retryErr := q.setRetry(task, err, elapsed); retryErr != nil {
				slog.Error("task.retry_failed",
					"id", task.ID,
					"preset", task.preset,
					"retry_times", task.retryTimes,
					"retry_max", task.retryMax,
					"error", err,
					"retry_error", retryErr,
				)
			}
			return
		}

		if task.retryOn && task.retryTimes >= task.retryMax {
			slog.Error("task.exhausted",
				"id", task.ID,
				"preset", task.preset,
				"retry_times", task.retryTimes,
				"retry_max", task.retryMax,
				"error", err,
				"elapsed_ms", elapsed.Milliseconds(),
			)
		} else {
			slog.Error("task.failed",
				"id", task.ID,
				"preset", task.preset,
				"error", err,
				"elapsed_ms", elapsed.Milliseconds(),
			)
		}
	} else {

		slog.Info("task.completed",
			"id", task.ID,
			"preset", task.preset,
			"retry_times", task.retryTimes,
			"elapsed_ms", elapsed.Milliseconds(),
		)

		// * Callback after task completion
		if task.callback != nil {
			go task.callback(task.ID)
		}
	}
}

func (q *Queue) setRetry(task *task, err error, elapsed time.Duration) error {
	slog.Warn("task.retrying",
		"id", task.ID,
		"preset", task.preset,
		"retry_times", task.retryTimes,
		"retry_max", task.retryMax,
		"error", err,
		"elapsed_ms", elapsed.Milliseconds(),
	)

	task.retryTimes++
	task.priority = PriorityRetry
	task.startAt = time.Now()

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return fmt.Errorf("queue closed, cannot retry")
	}

	return q.pending.Push(task)
}

func (q *Queue) Enqueue(ctx context.Context, presetName string, action func(ctx context.Context) error, options ...EnqueueOption) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	config := &enqueueConfig{
		timeout: q.config.getQueueTimeout(presetName),
	}
	for _, option := range options {
		option(config)
	}

	if config.taskID == "" {
		config.taskID = generateUUID()
	}

	var retryMax int
	if config.retryOn {
		retryMax = 3
		if config.retryMax != nil {
			retryMax = *config.retryMax
		}
	}

	task := &task{
		ID:       config.taskID,
		preset:   presetName,
		priority: q.config.Preset[presetName].Priority,
		action:   action,
		timeout:  config.timeout,
		callback: config.callback,
		startAt:  time.Now(),
		retryOn:  config.retryOn,
		retryMax: retryMax,
	}

	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return "", fmt.Errorf("enqueue failed: scheduler is closed")
	}
	err := q.pending.Push(task)
	q.mu.Unlock()

	if err != nil {
		return "", err
	}

	return task.ID, nil
}

func (q *Queue) Shutdown(ctx context.Context) error {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()

	q.pending.Close()

	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		if q.cancel != nil {
			q.cancel()
		}
		return fmt.Errorf("shutdown timeout: %d tasks remaining", q.pending.Len())
	}

	return nil
}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	buf := make([]byte, 36)
	hex.Encode(buf[0:8], b[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], b[10:16])
	return string(buf)
}
