package goQueue

import (
	"container/heap"
	"context"
	"runtime"
	"sync"
)

type Queue struct {
	config *Config
	heap   taskHeap
	cond   *sync.Cond
	wg     sync.WaitGroup
	mu     sync.Mutex
	closed bool
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
		config: newConfig,
		heap:   make(taskHeap, 0),
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *Queue) Start(ctx context.Context) {
	for i := 0; i < q.config.Workers; i++ {
		q.wg.Add(1)
		go q.worker(ctx)
	}
}

func (q *Queue) worker(ctx context.Context) {
	defer q.wg.Done()
	for {
		q.mu.Lock()
		for q.heap.Len() == 0 && !q.closed {
			q.cond.Wait()
		}

		if q.closed && q.heap.Len() == 0 {
			q.mu.Unlock()
			return
		}

		task := heap.Pop(&q.heap).(*Task)
		q.mu.Unlock()

		select {
		case <-ctx.Done():
			return
		default:
			task.Action(ctx)
		}
	}
}

func (q *Queue) Enqueue(presetName string, action func(ctx context.Context) error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return
	}

	heap.Push(&q.heap, &Task{
		Action:   action,
		Priority: q.config.getPresetPriority(presetName),
	})
	q.cond.Signal()
}

func (q *Queue) Shutdown() {
	q.mu.Lock()
	q.closed = true
	q.cond.Broadcast()
	q.mu.Unlock()

	q.wg.Wait()
}
