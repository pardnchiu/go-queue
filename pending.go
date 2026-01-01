package goQueue

import (
	"container/heap"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type pending struct {
	mu        sync.Mutex
	cond      *sync.Cond
	heap      *taskHeap
	size      int
	state     *atomic.Uint32
	promotion map[Priority]promotion
}

type promotion struct {
	After time.Duration
	To    Priority
}

type promotionTask struct {
	taskID string
	from   Priority
	to     Priority
}

func (c *Config) getPromotion() map[Priority]promotion {
	timeout := c.Timeout
	return map[Priority]promotion{
		PriorityLow: {
			After: min(max(timeout, 30*time.Second), 120*time.Second),
			To:    PriorityNormal,
		},
		PriorityNormal: {
			After: min(max(timeout*2, 30*time.Second), 120*time.Second),
			To:    PriorityHigh,
		},
	}
}

func newPending(workers, size int, promotion map[Priority]promotion, queueState *atomic.Uint32) *pending {
	minCap := max(16, min(size/8, size/workers))
	h := &taskHeap{
		minCap: minCap,
	}
	heap.Init(h)

	newPending := &pending{
		heap:      h,
		size:      size,
		promotion: promotion,
		state:     queueState,
	}
	newPending.cond = sync.NewCond(&newPending.mu)
	return newPending
}

func (p *pending) Push(t *task) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	current := queueState(p.state.Load())
	if current == stateClosed {
		return fmt.Errorf("staging queue is closed")
	}

	if p.heap.Len() >= p.size {
		return fmt.Errorf("staging queue is full")
	}

	heap.Push(p.heap, t)
	p.cond.Signal()
	return nil
}

func (p *pending) Pop() (*task, []promotionTask, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		state := queueState(p.state.Load())
		if state == stateClosed && p.heap.Len() == 0 {
			return nil, nil, false
		}

		events := p.promoteLocked()

		if p.heap.Len() > 0 {
			task := heap.Pop(p.heap).(*task)
			return task, events, true
		}

		if state == stateClosed {
			return nil, nil, false
		}

		p.cond.Wait()
	}
}

func (p *pending) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.heap.Len()
}

func (p *pending) promoteLocked() []promotionTask {
	var events []promotionTask
	now := time.Now()

	tasks := p.heap.tasks
	for i := p.heap.Len() - 1; i >= 0; i-- {
		t := tasks[i]
		rule, ok := p.promotion[t.priority]
		if !ok || now.Sub(t.startAt) < rule.After || rule.To >= t.priority {
			continue
		}
		events = append(events, promotionTask{
			taskID: t.ID,
			from:   t.priority,
			to:     rule.To,
		})
		t.priority = rule.To
		heap.Fix(p.heap, i)
	}
	return events
}

func (p *pending) State() queueState {
	return queueState(p.state.Load())
}

func (p *pending) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cond.Broadcast()
}
