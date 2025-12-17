package goQueue

import (
	"context"
	"sync"
)

type Task struct {
	Action func(ctx context.Context) error
}

type Queue struct {
	tasks chan *Task
	wg    sync.WaitGroup
}

func New() *Queue {
	return &Queue{
		tasks: make(chan *Task, 64),
	}
}

func (q *Queue) Start(ctx context.Context) {
	q.wg.Add(1)
	go q.worker(ctx)
}

func (q *Queue) worker(ctx context.Context) {
	defer q.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-q.tasks:
			if !ok {
				return
			}

			task.Action(ctx)
		}
	}
}

func (q *Queue) Enqueue(action func(ctx context.Context) error) {
	q.tasks <- &Task{Action: action}
}

func (q *Queue) Shutdown() {
	close(q.tasks)
	q.wg.Wait()
}
