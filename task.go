package goQueue

import (
	"context"
	"time"
)

const (
	taskHeapMinCap = 16
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

type taskHeap []*task

func (h taskHeap) Len() int {
	return len(h)
}

func (h taskHeap) Less(i, j int) bool {
	if h[i].priority != h[j].priority {
		return h[i].priority < h[j].priority
	}
	return h[i].startAt.Before(h[j].startAt)
}

func (h taskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *taskHeap) Push(x interface{}) {
	*h = append(*h, x.(*task))
}

func (h *taskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	task := old[n-1]
	old[n-1] = nil
	*h = old[0 : n-1]

	length := len(*h)
	capacity := cap(*h)
	if capacity > taskHeapMinCap && length < capacity/4 {
		newCap := capacity / 2
		if newCap < taskHeapMinCap {
			newCap = taskHeapMinCap
		}
		shrunk := make(taskHeap, length, newCap)
		copy(shrunk, *h)
		*h = shrunk
	}

	return task
}
