package goQueue

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Queue struct {
	config  *Config
	pending *pending
	ctx     context.Context
	wg      sync.WaitGroup
	mu      sync.RWMutex
	closed  bool
}

func New(config *Config) *Queue {
	worker := runtime.NumCPU() * 2
	newConfig := &Config{
		Workers:  worker,
		Size:     worker * 64,
		Timeout:  30,
		Priority: priorityNormal,
		Preset:   make(map[string]PresetConfig),
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
		if config.Priority != 0 {
			newConfig.Priority = config.Priority
		}
		if config.Preset != nil {
			newConfig.Preset = make(map[string]PresetConfig)
			for k, v := range config.Preset {
				newConfig.Preset[k] = v
			}
		}
	}
	q := &Queue{
		config:  newConfig,
		pending: newPending(newConfig.Size),
	}
	return q
}

func (q *Queue) Start(ctx context.Context) {
	q.ctx = ctx
	for i := 0; i < q.config.Workers; i++ {
		q.wg.Add(1)
		go q.worker(ctx)
	}
}

func (q *Queue) worker(ctx context.Context) {
	defer q.wg.Done()

	for {
		task, ok := q.pending.Pop()

		if !ok {
			return
		}

		q.execute(task)
	}
}

func (q *Queue) execute(task *task) {
	ctx, cancel := context.WithTimeout(q.ctx, task.timeout)
	defer cancel()

	errChan := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic: %v", r)
				return
			}
		}()
		errChan <- task.action(ctx)
	}()

	var err error
	select {
	case err = <-errChan:
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			err = fmt.Errorf("task timeout after %s", task.timeout)
		} else {
			err = ctx.Err()
		}
		go func() {
			<-errChan
		}()
	}

	if err != nil {
		fmt.Printf("Task %s failed: %v\n", task.ID, err)
	} else {
		fmt.Printf("Task %s completed successfully\n", task.ID)
	}
}

func (q *Queue) Enqueue(presetName string, action func(ctx context.Context) error) (string, error) {
	q.mu.RLock()
	if q.closed {
		q.mu.RUnlock()
		return "", fmt.Errorf("enqueue failed: scheduler is closed")
	}
	q.mu.RUnlock()

	task := &task{
		ID:       generateUUID(),
		preset:   presetName,
		priority: q.config.getPresetPriority(presetName),
		action:   action,
		timeout:  q.config.getQueueTimeout(presetName),
		startAt:  time.Now(),
	}

	if err := q.pending.Push(task); err != nil {
		return "", err
	}
	return task.ID, nil
}

func (q *Queue) Shutdown() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()

	q.pending.Close()

	q.wg.Wait()
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
