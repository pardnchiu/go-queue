package goQueue

import (
	"container/heap"
	"context"
)

type Task struct {
	Action   func(ctx context.Context) error
	Priority priority
}

type taskHeap []*Task

func (h taskHeap) Len() int {
	return len(h)
}

func (h taskHeap) Less(i, j int) bool {
	return h[i].Priority < h[j].Priority
}

func (h taskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *taskHeap) Push(x any) {
	*h = append(*h, x.(*Task))
}

func (h *taskHeap) Pop() any {
	oldHeap := *h
	oldLen := len(oldHeap)

	task := oldHeap[oldLen-1]
	oldHeap[oldLen-1] = nil // * 避免 memory leak
	*h = oldHeap[:oldLen-1]

	return task
}

var _ heap.Interface = (*taskHeap)(nil)
