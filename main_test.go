package goQueue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestBasicEnqueue(t *testing.T) {
	queue := New()
	ctx := context.Background()
	queue.Start(ctx)

	var executed atomic.Bool

	queue.Enqueue(func(ctx context.Context) error {
		executed.Store(true)
		return nil
	})

	queue.Shutdown()

	if !executed.Load() {
		t.Error("task was not executed")
	}
}

func TestMultipleTasks(t *testing.T) {
	queue := New()
	ctx := context.Background()
	queue.Start(ctx)

	var count atomic.Int32

	for i := 0; i < 10; i++ {
		queue.Enqueue(func(ctx context.Context) error {
			count.Add(1)
			time.Sleep(5 * time.Millisecond)
			return nil
		})
	}

	queue.Shutdown()

	if count.Load() != 10 {
		t.Errorf("expected 10 tasks executed, got %d", count.Load())
	}
}

func TestContextCancellation(t *testing.T) {
	queue := New()
	ctx, cancel := context.WithCancel(context.Background())
	queue.Start(ctx)

	var started atomic.Bool
	var completed atomic.Bool

	queue.Enqueue(func(ctx context.Context) error {
		started.Store(true)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			completed.Store(true)
			return nil
		}
	})

	time.Sleep(50 * time.Millisecond)
	cancel()
	queue.Shutdown()

	if !started.Load() {
		t.Error("task should have started")
	}
	if completed.Load() {
		t.Error("task should have been cancelled")
	}
}

func TestShutdownWaitsForCompletion(t *testing.T) {
	queue := New()
	ctx := context.Background()
	queue.Start(ctx)

	var completed atomic.Bool

	queue.Enqueue(func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		completed.Store(true)
		return nil
	})

	queue.Shutdown()

	if !completed.Load() {
		t.Error("shutdown should wait for task completion")
	}
}

func TestBufferedChannel(t *testing.T) {
	queue := New()
	ctx := context.Background()
	queue.Start(ctx)

	var count atomic.Int32

	for i := 0; i < 100; i++ {
		queue.Enqueue(func(ctx context.Context) error {
			count.Add(1)
			return nil
		})
	}

	queue.Shutdown()

	if count.Load() != 100 {
		t.Errorf("expected 100 tasks, got %d", count.Load())
	}
}
