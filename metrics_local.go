package main

import (
	"sync"
	"time"
)

// localHandlerMetrics holds counters per handler
type localHandlerMetrics struct {
	Count          uint64  // number of calls
	DurationSumSec float64 // total duration in seconds
	Panics         uint64  // number of panics
}

var (
	metricsMu sync.Mutex
	metrics   = make(map[string]*localHandlerMetrics) // key: handler name (Param)
)

// recordRequestMetrics updates local metrics for one handler call
func recordRequestMetrics(handler string, duration time.Duration, panicked bool) {
	metricsMu.Lock()
	defer metricsMu.Unlock()

	m, ok := metrics[handler]
	if !ok {
		m = &localHandlerMetrics{}
		metrics[handler] = m
	}
	m.Count++
	m.DurationSumSec += duration.Seconds()
	if panicked {
		m.Panics++
	}
}

// snapshotAndReset returns current metrics and resets them
func snapshotAndReset() map[string]localHandlerMetrics {
	metricsMu.Lock()
	defer metricsMu.Unlock()

	out := make(map[string]localHandlerMetrics, len(metrics))
	for k, v := range metrics {
		out[k] = *v
	}

	// reset
	metrics = make(map[string]*localHandlerMetrics)

	return out
}
