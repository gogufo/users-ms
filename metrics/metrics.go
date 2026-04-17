package metrics

import (
	"sync/atomic"
	"time"
)

var (
	requestsTotal     uint64
	requestsPanicked  uint64
	totalDurationNano uint64
)

// RecordRequestMetrics updates internal aggregated metrics.
// method  — string label ("/Reverse/Do", "GET:info", …)
// duration — request duration
// panicked — whether handler panicked or returned error
func RecordRequestMetrics(method string, duration time.Duration, panicked bool) {
	// Count total requests
	atomic.AddUint64(&requestsTotal, 1)

	// Sum duration
	atomic.AddUint64(&totalDurationNano, uint64(duration.Nanoseconds()))

	// Count panics/errors
	if panicked {
		atomic.AddUint64(&requestsPanicked, 1)
	}
}

// Stats returns current aggregated snapshot.

func Stats() map[string]interface{} {
	total := atomic.LoadUint64(&requestsTotal)
	pan := atomic.LoadUint64(&requestsPanicked)
	dur := atomic.LoadUint64(&totalDurationNano)

	avg := float64(0)
	if total > 0 {
		avg = float64(dur) / float64(total) / 1e6 // ms
	}

	return map[string]interface{}{
		"total_requests": total,
		"panics":         pan,
		"avg_ms":         avg,
	}
}
