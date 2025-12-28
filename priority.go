package goQueue

import (
	"time"
)

type Priority int

const (
	PriorityImmediate Priority = iota
	PriorityHigh
	PriorityRetry
	PriorityNormal
	PriorityLow
)

func (c *Config) getQueueTimeout(name string) time.Duration {
	timeout := time.Duration(c.Timeout)
	if config, ok := c.Preset[name]; ok && config.Timeout > 0 {
		timeout = time.Duration(config.Timeout)
	}

	var dur time.Duration
	switch c.Preset[name].Priority {
	case PriorityImmediate:
		dur = time.Duration(timeout / 4)
	case PriorityHigh:
		dur = time.Duration(timeout / 2)
	case PriorityRetry:
		dur = time.Duration(timeout / 2)
	case PriorityLow:
		dur = time.Duration(timeout * 2)
	default:
		dur = time.Duration(timeout)
	}

	// 限制 dur 最小 15 秒，最大 120 秒
	if dur < 15*time.Second {
		dur = 15 * time.Second
	}
	if dur > 120*time.Second {
		dur = 120 * time.Second
	}

	return dur
}
