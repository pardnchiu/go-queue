package goQueue

import (
	"strings"
	"time"
)

type priority int

const (
	priorityImmediate priority = iota
	priorityHigh
	priorityNormal
	priorityLow
)

func (c *Config) getPresetPriority(name string) priority {
	if config, ok := c.Preset[name]; ok {
		switch strings.ToLower(config.Priority) {
		case "immediate":
			return priorityImmediate
		case "high":
			return priorityHigh
		case "normal":
			return priorityNormal
		case "low":
			return priorityLow
		default:
			return priorityNormal
		}
	}
	return priorityNormal
}

func (c *Config) getQueueTimeout(name string) time.Duration {
	timeout := c.Timeout
	if config, ok := c.Preset[name]; ok && config.Timeout > 0 {
		timeout = config.Timeout
	}

	var sec int
	switch c.getPresetPriority(name) {
	case priorityImmediate:
		sec = timeout / 4
	case priorityHigh:
		sec = timeout / 2
	case priorityNormal:
		sec = timeout
	case priorityLow:
		sec = timeout * 2
	default:
		sec = timeout
	}

	return time.Duration(min(max(sec, 15), 120)) * time.Second
}
