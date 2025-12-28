package goQueue

import "time"

type enqueueConfig struct {
	taskID   string
	timeout  time.Duration
	callback func(string)
	retryOn  bool
	retryMax *int
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

func WithCallback(fn func(id string)) EnqueueOption {
	return func(c *enqueueConfig) {
		c.callback = fn
	}
}

func WithRetry(retryMax ...int) EnqueueOption {
	return func(c *enqueueConfig) {
		c.retryOn = true
		if len(retryMax) > 0 {
			c.retryMax = &retryMax[0]
		} else {
			c.retryMax = nil
		}
	}
}
