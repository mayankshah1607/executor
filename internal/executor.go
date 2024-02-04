package internal

import (
	"errors"
	"sync/atomic"
)

// Task is the function signature of what is excuted by the workers.
type Task func()

type Executor struct {
	queue     chan Task
	queueSize int

	// Atomic counters to keep track of internal stats.
	// Mostly used for testing, but can also be used for other purposes.
	totalWorkers  int32 // total number of workers that are spawned.
	activeWorkers int32 // total number of workers that are currently doing work.
}

// NewExecutor returns a new async task executor with the provided queueSize.
// If at anytime more than `queueSize` number of tasks are provided, adding to the executor becomes a blocking operation.
func NewExecutor(queueSize int) *Executor {
	return &Executor{
		queueSize: queueSize,
		queue:     make(chan Task, queueSize),
	}
}

var MaxQueueCapacity = errors.New("Queue is operating at maximum capacity")

// AddTask enqueues a new task to the executor.
// This call will be a blocking operation if the size of the queue is exceeded.
func (e *Executor) AddTask(t Task) error {
	if len(e.queue) == e.queueSize {
		return MaxQueueCapacity
	}
	e.queue <- t
	return nil
}

// Stop stops all the go routines in this executor.
// Not calling Stop may result in leaking goroutines.
func (e *Executor) Stop() {
	close(e.queue)
}

// Spawn the specified number of workers.
// Caller is responsible for calling 'Stop' in order to prevent leaks.
func (e *Executor) Spawn(n int) {
	// Start 'n' workers.
	for i := 0; i < n; i++ {
		go func() {
			atomic.AddInt32(&e.totalWorkers, 1)
			defer atomic.AddInt32(&e.totalWorkers, -1)

			// Keep reading from the queue until the channel is closed.
			for task := range e.queue {
				atomic.AddInt32(&e.activeWorkers, 1)
				task()
				atomic.AddInt32(&e.activeWorkers, -1)
			}
		}()
	}
}

// TotalWorkers returns the total number of workers in this executor.
func (e *Executor) TotalWorkers() int32 {
	return atomic.LoadInt32(&e.totalWorkers)
}

// ActiveWorkers returns the total number of active workers at any given moment in this executor.
func (e *Executor) ActiveWorkers() int32 {
	return atomic.LoadInt32(&e.activeWorkers)
}
