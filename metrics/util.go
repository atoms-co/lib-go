package metrics

import (
	"context"
	"strings"
	"sync"
)

// TrackedGauge is a Gauge wrapper that dynamically tracks which tags have been used. It allows
// a reset of the involved values. Useful for gauges that capture transient values that should be
// reported as zero when no longer present.
type TrackedGauge struct {
	g        Gauge
	observed map[string][]Tag
	mu       sync.RWMutex
}

func NewTrackedGauge(g Gauge) *TrackedGauge {
	return &TrackedGauge{g: g, observed: map[string][]Tag{}}
}

// Set the gauge to the given arbitrary value values is a map of all the tag keys with their
// corresponding values.
func (g *TrackedGauge) Set(ctx context.Context, value float64, tags ...Tag) {
	hash := hashTags(tags)
	g.mu.RLock()
	_, ok := g.observed[hash]
	g.mu.RUnlock()
	if !ok {
		g.mu.Lock()
		g.observed[hash] = tags
		g.mu.Unlock()
	}

	g.g.Set(ctx, value, tags...)
}

// Reset resets all values with the given tag prefix. For performance, we consider tags ordered and require
// consistency across Set and Reset.
func (g *TrackedGauge) Reset(ctx context.Context, tags ...Tag) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for k, used := range g.observed {
		if !isPrefix(tags, used) {
			continue
		}
		g.g.Set(ctx, 0, used...)
		delete(g.observed, k)
	}
}

func isPrefix(tags, used []Tag) bool {
	if len(used) < len(tags) {
		return false
	}
	for i := 0; i < len(tags); i++ {
		if tags[i] != used[i] {
			return false
		}
	}
	return true
}

func hashTags(tags []Tag) string {
	var sb strings.Builder
	for _, t := range tags {
		sb.WriteString(t.String())
		sb.WriteString("!")
	}
	return sb.String()
}
