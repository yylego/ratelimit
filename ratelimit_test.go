package ratelimit

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLimiter_Allow(t *testing.T) {
	lm := NewLimiter(100)

	accepted := 0
	for i := 0; i < 200; i++ {
		if lm.Allow() {
			accepted++
		}
	}
	require.Equal(t, 100, accepted)
}

func TestLimiter_SlidingWindowResets(t *testing.T) {
	lm := NewLimiter(10)

	for i := 0; i < 10; i++ {
		require.True(t, lm.Allow())
	}
	require.False(t, lm.Allow())

	time.Sleep(time.Second + 50*time.Millisecond)

	require.True(t, lm.Allow())
}

func TestGroup_SingleKeyOverLimit(t *testing.T) {
	gp := NewGroup(1000, time.Second)

	var accepted atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			if gp.Allow("hot_name") {
				accepted.Add(1)
			}
		}
	}()
	wg.Wait()

	t.Log("accepted:", accepted.Load())
	require.True(t, accepted.Load() <= 1000)
	require.True(t, accepted.Load() >= 900)
}

func TestGroup_MultiKeyIsolation(t *testing.T) {
	gp := NewGroup(500, time.Second)

	var acceptedA, acceptedB atomic.Int64
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			if gp.Allow("name_a") {
				acceptedA.Add(1)
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			if gp.Allow("name_b") {
				acceptedB.Add(1)
			}
		}
	}()
	wg.Wait()

	t.Log("name_a accepted:", acceptedA.Load())
	t.Log("name_b accepted:", acceptedB.Load())

	require.True(t, acceptedA.Load() <= 500)
	require.True(t, acceptedA.Load() >= 400)
	require.True(t, acceptedB.Load() <= 500)
	require.True(t, acceptedB.Load() >= 400)
}

func TestGroup_AutoEviction(t *testing.T) {
	gp := NewGroup(100, 100*time.Millisecond)

	gp.Allow("temp_name")

	gp.mu.Lock()
	require.Equal(t, 1, len(gp.dataItems))
	require.Equal(t, 1, gp.expireHeap.Len()) // 1:1 with map
	gp.mu.Unlock()

	time.Sleep(200 * time.Millisecond)

	// trigger eviction via Allow on a different key
	gp.Allow("ping")

	gp.mu.Lock()
	_, exists := gp.dataItems["temp_name"]
	require.False(t, exists)
	require.Equal(t, 1, len(gp.dataItems))   // "ping" remains
	require.Equal(t, 1, gp.expireHeap.Len()) // 1:1 with map
	gp.mu.Unlock()
}
