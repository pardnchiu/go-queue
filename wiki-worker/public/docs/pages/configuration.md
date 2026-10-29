# Configuration

This page covers every queue-level `Config` field, its default, and how zero values resolve. go-queue reads no environment variables or config files; all settings go into `core.New` as a struct.

## Config

```go
type Config struct {
	Workers int
	Size    int
	Timeout time.Duration
	Preset  map[string]PresetConfig
}
```

| Field | Default When Zero | Description |
|---|---|---|
| `Workers` | `runtime.NumCPU() * 2` | Number of concurrent workers |
| `Size` | `Workers * 64` (using the default Workers) | Heap capacity, retries included; `Enqueue` errors once full |
| `Timeout` | `30 * time.Second` | Global base timeout; also sets promotion thresholds |
| `Preset` | empty map | Preset name → `PresetConfig`; `New` copies it, so later edits to your map do not affect the queue |

## Zero-Value Resolution

| Input | Result |
|---|---|
| `New(nil)` | Every default |
| Only `Workers: 1` | `Size` stays `NumCPU * 2 * 64`: the `Size` default derives from `runtime.NumCPU() * 2`, not from the `Workers` you pass (`core/new.go:44`) |
| Negative values | Not rejected; non-zero values are used as given |

Named presets are covered in [Presets](/configuration-presets); per-task adjustments in [Enqueue Options](/api-reference-options).
