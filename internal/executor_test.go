package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Note that these tests are intentionally slow because of the time.Sleep calls.
// This is so that we give the executor adequate time to clean-up/spawn goroutines so
// that we can reliably make assertions about the internal state.
func TestExecutor(t *testing.T) {
	t.Run("test stop", func(t *testing.T) {
		e := NewExecutor(10)
		e.Spawn(3)
		time.Sleep(2 * time.Second) // wait for the workers to come up. TODO: use retry here.
		assert.Equal(t, int32(3), e.TotalWorkers())
		assert.Equal(t, int32(0), e.ActiveWorkers())

		e.Stop()
		time.Sleep(2 * time.Second) // wait for the workers to exit. TODO: use retry here.
		assert.Equal(t, int32(0), e.TotalWorkers())
		assert.Equal(t, int32(0), e.ActiveWorkers())
	})
	t.Run("test workers", func(t *testing.T) {
		e := NewExecutor(10)
		e.Spawn(3)
		time.Sleep(2 * time.Second) // wait for the workers to come up. TODO: use retry here.

		// Add 3 tasks.
		for i := 0; i < 3; i++ {
			e.AddTask(func() { time.Sleep(time.Second * 10) })
		}
		time.Sleep(2 * time.Second) // wait for the workers to pick up tasks from the queue.
		assert.Equal(t, int32(3), e.ActiveWorkers())
		assert.Equal(t, int32(3), e.TotalWorkers())

		time.Sleep(10 * time.Second) // wait for workers to finish their work.
		assert.Equal(t, int32(0), e.ActiveWorkers())
		assert.Equal(t, int32(3), e.TotalWorkers())
		e.Stop()
		time.Sleep(2 * time.Second) // wait for the workers to exit. TODO: use retry here.
		assert.Equal(t, int32(0), e.TotalWorkers())
		assert.Equal(t, int32(0), e.ActiveWorkers())
	})
}
