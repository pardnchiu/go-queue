package goQueue

import "time"

type enqueueConfig struct {
	taskID   string
	timeout  time.Duration
	callback func(string, error)
}

type EnqueueOption func(*enqueueConfig)

func WithTaskID(id string) EnqueueOption {
	return func(c *enqueueConfig) {
		c.taskID = id
	}
}

func WithTimeout(d time.Duration) EnqueueOption {
	return func(c *enqueueConfig) {
		c.timeout = d
	}
}

func WithCallback(fn func(id string, err error)) EnqueueOption {
	return func(c *enqueueConfig) {
		c.callback = fn
	}
}
