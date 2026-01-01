package goQueue

import (
	"context"
	"time"
)

const (
	taskHeapMinCapRatio = 8
)

type task struct {
	ID         string
	preset     string
	priority   Priority
	action     func(ctx context.Context) error
	timeout    time.Duration
	callback   func(id string)
	startAt    time.Time
	retryOn    bool
	retryMax   int
	retryTimes int
}

type taskHeaps []*task

type taskHeap struct {
	tasks  taskHeaps
	minCap int
}

func (h *taskHeap) Len() int {
	return len(h.tasks)
}

func (h *taskHeap) Less(i, j int) bool {
	if h.tasks[i].priority != h.tasks[j].priority {
		return h.tasks[i].priority < h.tasks[j].priority
	}
	return h.tasks[i].startAt.Before(h.tasks[j].startAt)
}

func (h *taskHeap) Swap(i, j int) {
	h.tasks[i], h.tasks[j] = h.tasks[j], h.tasks[i]
}

func (h *taskHeap) Push(x interface{}) {
	h.tasks = append(h.tasks, x.(*task))
}

func (h *taskHeap) Pop() interface{} {
	old := h.tasks
	n := len(old)
	task := old[n-1]
	old[n-1] = nil
	h.tasks = old[0 : n-1]

	length := len(h.tasks)
	capacity := cap(h.tasks)
	if capacity > h.minCap*4 && length < capacity/taskHeapMinCapRatio {
		newCap := max(capacity/4, h.minCap)
		shrunk := make(taskHeaps, length, newCap)
		copy(shrunk, h.tasks)
		h.tasks = shrunk
	}

	return task
}
