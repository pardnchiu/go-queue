package goQueue

import (
	"container/heap"
	"fmt"
	"sync"
)

type pending struct {
	mu     sync.Mutex
	cond   *sync.Cond
	heap   *taskHeap
	size   int
	closed bool
}

func newPending(size int) *pending {
	newPending := &pending{
		heap: &taskHeap{},
		size: size,
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

func (p *pending) Pop() (*task, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		if p.closed && p.heap.Len() == 0 {
			return nil, false
		}

		if p.heap.Len() > 0 {
			task := heap.Pop(p.heap).(*task)
			return task, true
		}
		if p.closed {
			return nil, false
		}

		p.cond.Wait()
	}
}

func (p *pending) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.heap.Len()
}

func (p *pending) PendingByPriority() map[priority]int {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := make(map[priority]int)
	for _, e := range *p.heap {
		result[e.priority]++
	}
	return result
}

func (p *pending) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	p.cond.Broadcast()
}
