package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/mayankshah1607/task-executor/internal"
	"github.com/rs/zerolog"
)

// Service uses an Executor to schedule async tasks.
// It contains the core business logic of our application.
// It also keeps track of the waiting and running tasks.
type Service struct {
	executor *internal.Executor

	// notify this channel whenever a task has started.
	startedNotifs chan string
	// notify this channel whenever a task has completed.
	completionNotifs chan string

	// internal execution data.
	// these structures will be updated by only 1 goroutine.
	waiting map[string]struct{}
	running map[string]struct{}
	lock    sync.RWMutex // the above maps may be read/updated concurrently, so we need to synchronize access.

	ctx context.Context
	log zerolog.Logger
}

// NewService returns a new service.
func NewService(log zerolog.Logger, executor *internal.Executor) *Service {
	return &Service{
		log:              log,
		executor:         executor,
		startedNotifs:    make(chan string),
		completionNotifs: make(chan string),
		waiting:          make(map[string]struct{}),
		running:          make(map[string]struct{}),
	}
}

// AddTasks adds a list of given tasks to the executor.
func (s *Service) AddTasks(tasks map[string]int32) error {
	if s.ctx == nil {
		return fmt.Errorf("Run not called")
	}

	// Add each task to execution.
	for id, dur := range tasks {
		id := id
		t := time.Second * time.Duration(dur)

		// (Critical section) Check if this task is already in our executor?
		{
			s.lock.Lock()
			_, waiting := s.waiting[id]
			_, running := s.running[id]
			if waiting || running {
				// in execution, unlock and move to next.
				s.lock.Unlock()
				continue
			}
			s.waiting[id] = struct{}{}
			s.lock.Unlock()
		}

		log := s.log.With().Str("id", id).Logger()
		// Add task to the executor.
		if err := s.executor.AddTask(func() {
			log.Info().Msg("Started task")

			s.startedNotifs <- id // notify that this task is started.
			defer func() {
				s.completionNotifs <- id // notify when this task is completed.
				log.Info().Msg("Finished task")
			}()

			t := time.NewTimer(t)
			defer t.Stop()

			// Wait for the timer or context.
			for {
				select {
				case <-s.ctx.Done():
					return
				case <-t.C:
					return
				}
			}
		}); err != nil {
			return err
		}
	}
	return nil
}

// Returns: Waiting, Running, Error
func (s *Service) GetStatistics() ([]string, []string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	waiting := make([]string, 0, len(s.waiting))
	running := make([]string, 0, len(s.running))

	for k := range s.waiting {
		waiting = append(waiting, k)
	}
	for k := range s.running {
		running = append(running, k)
	}
	sort.Strings(waiting)
	sort.Strings(running)
	return waiting, running, nil
}

// Run updates the statistics internally whenever a task is started/completed.
// This should be called only once from within a go routine, or else it will return an error.
func (s *Service) Run(ctx context.Context) error {
	if s.ctx != nil {
		return fmt.Errorf("already running")
	}
	s.ctx = ctx
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case id := <-s.startedNotifs:
			// A new task was started, update internal state.
			s.lock.Lock()
			delete(s.waiting, id)
			s.running[id] = struct{}{}
			s.lock.Unlock()
		case id := <-s.completionNotifs:
			// A task was completed, update internal state.
			s.lock.Lock()
			delete(s.running, id)
			s.lock.Unlock()
		}
	}
}
