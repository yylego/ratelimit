package ratelimit

import (
	"math"
	"sync"
	"time"

	"github.com/yylego/heapx"
)

// expireItem is the value stored in the heap, tracking key and access time
type expireItem struct {
	key        string
	accessNano int64
	lm         *Limiter
}

// Group is a concurrent-safe per-key rate limiter with heap-based auto-eviction
// Map and heap are 1:1 — each key has exactly one heap node, no duplicates
type Group struct {
	dataItems         map[string]*heapx.Node[expireItem]
	threshold         int
	windowNano        int64
	mu                *sync.Mutex
	expireHeap        *heapx.Heap[expireItem]
	sweepBatch        int           // max evictions per sweep
	sweepInBackground bool          // when true, Allow() skips sweep
	sweepDone         chan struct{} // close to stop background goroutine
}

// NewGroup creates a per-key rate limiter group with the given threshold and window
// No background goroutine — eviction happens lazily via min-heap on each Allow() call
func NewGroup(threshold int, window time.Duration) *Group {
	return &Group{
		dataItems:  make(map[string]*heapx.Node[expireItem]),
		threshold:  threshold,
		windowNano: window.Nanoseconds(),
		mu:         &sync.Mutex{},
		expireHeap: heapx.New[expireItem](func(a, b expireItem) bool {
			return a.accessNano < b.accessNano
		}),
		sweepBatch: math.MaxInt, // default: no cap, clean all expired
	}
}

// SetSweepBatch sets the maximum evictions per sweep call
func (G *Group) SetSweepBatch(n int) {
	G.mu.Lock()
	G.sweepBatch = n
	G.mu.Unlock()
}

// Allow checks if the request to the given key is allowed
func (G *Group) Allow(key string) bool {
	G.mu.Lock()
	now := time.Now().UnixNano()

	if !G.sweepInBackground {
		G.sweep(now)
	}

	node, ok := G.dataItems[key]
	if !ok {
		node = G.expireHeap.Push(expireItem{
			key:        key,
			accessNano: now,
			lm:         NewLimiterWithWindow(G.threshold, time.Duration(G.windowNano)),
		})
		G.dataItems[key] = node
	} else {
		node.Value.accessNano = now
		G.expireHeap.Fix(node)
	}
	lm := node.Value.lm
	G.mu.Unlock()

	// lm has its own mutex, no need to hold group lock
	return lm.Allow()
}

// StartSweepGoroutine starts a background goroutine that sweeps expired keys on a fixed tick
// Once started, Allow() stops inline sweep — the background goroutine handles it
// Call CloseSweepGoroutine() to close and switch back to inline sweep
func (G *Group) StartSweepGoroutine(tick time.Duration) {
	G.mu.Lock()
	if G.sweepInBackground {
		G.mu.Unlock()
		return
	}
	G.sweepInBackground = true
	G.sweepDone = make(chan struct{})
	G.mu.Unlock()

	go func() {
		t := time.NewTicker(tick)
		defer t.Stop()
		for {
			select {
			case <-G.sweepDone:
				return
			case <-t.C:
				G.mu.Lock()
				G.sweep(time.Now().UnixNano())
				G.mu.Unlock()
			}
		}
	}()
}

// CloseSweepGoroutine closes the background sweep goroutine and switches back to inline sweep
func (G *Group) CloseSweepGoroutine() {
	G.mu.Lock()
	if G.sweepInBackground {
		close(G.sweepDone)
		G.sweepInBackground = false
	}
	G.mu.Unlock()
}

// KeysCount returns the count of active keys in the group
func (G *Group) KeysCount() int {
	G.mu.Lock()
	n := len(G.dataItems)
	G.mu.Unlock()
	return n
}

// sweep removes idle keys from the heap top, up to sweepBatch
func (G *Group) sweep(now int64) {
	cutoff := now - G.windowNano
	count := 0
	for G.expireHeap.Len() > 0 {
		if count >= G.sweepBatch {
			break
		}
		top := G.expireHeap.Peek()
		if top == nil || top.Value.accessNano > cutoff {
			break
		}
		G.expireHeap.Pop()
		delete(G.dataItems, top.Value.key)
		count++
	}
}
