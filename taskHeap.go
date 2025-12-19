package goQueue

import (
	"context"
	"time"
)

type task struct {
	ID       string
	preset   string // 自訂通道名稱
	priority priority
	action   func(ctx context.Context) error
	timeout  time.Duration
	startAt  time.Time
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
	*h = old[0 : n-1]
	return task
}
