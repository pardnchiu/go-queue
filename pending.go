package goQueue

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type pending struct {
	mu        sync.Mutex
	cond      *sync.Cond
	heap      *taskHeap
	size      int
	closed    bool
	promotion map[priority]promotion
}

type promotion struct {
	After time.Duration
	To    priority
}

type promotionTask struct {
	taskID string
	from   priority
	to     priority
}

func (c *Config) getPromotion() map[priority]promotion {
	timeout := time.Duration(c.Timeout) * time.Second
	return map[priority]promotion{
		priorityLow: {
			After: min(max(timeout, 30*time.Second), 120*time.Second),
			To:    priorityNormal,
		},
		priorityNormal: {
			After: min(max(timeout*2, 30*time.Second), 120*time.Second),
			To:    priorityHigh,
		},
	}
}

func newPending(size int, promotion map[priority]promotion) *pending {
	h := &taskHeap{}
	heap.Init(h)

	newPending := &pending{
		heap:      h,
		size:      size,
		promotion: promotion,
	}
	newPending.cond = sync.NewCond(&newPending.mu)
	return newPending
}

func (p *pending) Push(t *task) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("staging queue is disabled")
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
		if p.closed && p.heap.Len() == 0 {
			return nil, nil, false
		}

		events := p.promoteLocked()

		if p.heap.Len() > 0 {
			task := heap.Pop(p.heap).(*task)
			return task, events, true
		}
		if p.closed {
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

	for i := p.heap.Len() - 1; i >= 0; i-- {
		t := (*p.heap)[i]
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

func (p *pending) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	p.cond.Broadcast()
}
