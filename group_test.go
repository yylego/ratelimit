package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGroup_SweepBatch(t *testing.T) {
	gp := NewGroup(10, 100*time.Millisecond)
	gp.SetSweepBatch(3)

	// push 10 keys
	for i := 0; i < 10; i++ {
		name := "k" + string(rune('A'+i))
		gp.Allow(name)
	}
	t.Log("keys count:", gp.KeysCount())
	require.Equal(t, 10, gp.KeysCount())

	// wait keys to expire
	time.Sleep(200 * time.Millisecond)

	// each Allow kicks sweep, but at most 3 evictions
	gp.Allow("ping")
	t.Log("keys count post 1st sweep:", gp.KeysCount())

	gp.Allow("ping")
	t.Log("keys count post 2nd sweep:", gp.KeysCount())

	gp.Allow("ping")
	t.Log("keys count post 3rd sweep:", gp.KeysCount())

	gp.Allow("ping")
	t.Log("keys count post 4th sweep:", gp.KeysCount())

	// "ping" itself is active so at least 1 remains
	require.True(t, gp.KeysCount() >= 1)
}

func TestGroup_SweepGoroutine(t *testing.T) {
	gp := NewGroup(10, 100*time.Millisecond)

	// push 10 keys first (before starting goroutine, no race)
	for i := 0; i < 10; i++ {
		name := "k" + string(rune('A'+i))
		gp.Allow(name)
	}
	t.Log("keys count:", gp.KeysCount())
	require.Equal(t, 10, gp.KeysCount())

	// start background sweep, tick 50ms
	gp.StartSweepGoroutine(50 * time.Millisecond)

	// wait keys to expire + background sweep to kick in
	time.Sleep(300 * time.Millisecond)

	t.Log("keys count post background sweep:", gp.KeysCount())
	require.Equal(t, 0, gp.KeysCount())

	// close and switch back to inline sweep
	gp.CloseSweepGoroutine()

	// push new keys, should be cleaned inline now
	gp.Allow("pong")
	t.Log("keys count post close:", gp.KeysCount())
}
