package ratelimit

import (
	"sort"
	"sync"
	"time"
)

// Limiter is a sliding window rate limiter for a single key
// Counts requests in the configured time window and rejects when exceeding the threshold
type Limiter struct {
	timestamps []int64
	mu         *sync.Mutex
	threshold  int
	windowNano int64
}

// NewLimiter creates a single-key limiter with the given threshold and 1-second window
func NewLimiter(threshold int) *Limiter {
	return NewLimiterWithWindow(threshold, time.Second)
}

// NewLimiterWithWindow creates a single-key limiter with custom threshold and window duration
func NewLimiterWithWindow(threshold int, window time.Duration) *Limiter {
	return &Limiter{
		timestamps: make([]int64, 0),
		mu:         &sync.Mutex{},
		threshold:  threshold,
		windowNano: window.Nanoseconds(),
	}
}

// Allow checks if one more request is within the limit
func (lm *Limiter) Allow() bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	now := time.Now().UnixNano()
	cutoff := now - lm.windowNano

	// trim expired entries using binary search
	idx := sort.Search(len(lm.timestamps), func(i int) bool { return lm.timestamps[i] >= cutoff })
	if idx == len(lm.timestamps) {
		lm.timestamps = lm.timestamps[:0]
	} else {
		lm.timestamps = lm.timestamps[idx:]
	}

	if len(lm.timestamps) >= lm.threshold {
		return false
	}

	lm.timestamps = append(lm.timestamps, now)
	return true
}
