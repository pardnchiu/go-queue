# Presets

This page explains `PresetConfig`, how a preset sets a task's priority and base timeout, and what an undefined preset name resolves to.

## PresetConfig

```go
type PresetConfig struct {
	Priority Priority
	Timeout  time.Duration
}
```

| Field | Zero Value | Description |
|---|---|---|
| `Priority` | `PriorityImmediate` | Initial priority for tasks using this preset |
| `Timeout` | `0` | When `> 0`, replaces `Config.Timeout` as this preset's base timeout |

A preset name missing from the map (including `""`) resolves to the zero `PresetConfig{}`, which means `PriorityImmediate` plus the global base timeout.

## Example

```go
q := core.New(&core.Config{
	Workers: 8,
	Size:    1024,
	Timeout: 30 * time.Second,
	Preset: map[string]core.PresetConfig{
		"payment": {Priority: core.PriorityHigh},
		"email":   {Priority: core.PriorityNormal},
		"report":  {Priority: core.PriorityLow, Timeout: 60 * time.Second},
	},
})
```

| Preset | Priority | Effective Timeout | Promotion |
|---|---|---|---|
| `payment` | High | `30s ÷ 2 = 15s` | None |
| `email` | Normal | `30s` | To High after 60s waiting |
| `report` | Low | `min(60s × 2, 120s) = 120s` | To Normal after 30s waiting |
| `""` or undefined | Immediate | `max(30s ÷ 4, 15s) = 15s` | None |

Promotion thresholds read only `Config.Timeout`, never a preset's `Timeout`. The rules behind each column: [Priority Levels](/core-concepts-priority), [Timeout Derivation](/core-concepts-timeout), [Anti-Starvation Promotion](/core-concepts-promotion).
