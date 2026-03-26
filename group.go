package ratelimit

import (
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
	dataItems  map[string]*heapx.Node[expireItem]
	threshold  int
	windowNano int64
	mu         *sync.Mutex
	expireHeap *heapx.Heap[expireItem]
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
	}
}

// Allow checks if the request to the given key is allowed
func (G *Group) Allow(key string) bool {
	G.mu.Lock()
	now := time.Now().UnixNano()

	G.sweep(now)

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

// sweep removes idle keys from the heap top
func (G *Group) sweep(now int64) {
	cutoff := now - G.windowNano
	for G.expireHeap.Len() > 0 {
		top := G.expireHeap.Peek()
		if top == nil || top.Value.accessNano > cutoff {
			break
		}
		G.expireHeap.Pop()
		delete(G.dataItems, top.Value.key)
	}
}
