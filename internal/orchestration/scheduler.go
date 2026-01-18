package orchestration

import (
	"context"
	"fmt"
	"sync"
)

// Scheduler enqueues and runs tasks when their declared dependencies are
// satisfied. It doesn't know how to execute a task's workflow: an executor
// callback is provided by the caller.
type Scheduler struct {
	// dependents maps a node to the list of nodes that depend on it.
	dependents map[string][]string
	// inDegree tracks number of unmet dependencies per node.
	inDegree map[string]int
	// total nodes
	total int
}

// NewScheduler builds a Scheduler from a dependency graph as produced by
// BuildDependencyGraph (dependency -> dependents).
func NewScheduler(graph map[string][]string) *Scheduler {
	inDeg := map[string]int{}
	dependents := map[string][]string{}

	// ensure every node is present
	for n := range graph {
		inDeg[n] = 0
		dependents[n] = []string{}
	}

	// graph maps dependency -> dependents; compute inDegree counts
	for dep, adj := range graph {
		for _, v := range adj {
			inDeg[v]++
			// ensure the dependent exists in dependents map
			if _, ok := dependents[dep]; !ok {
				dependents[dep] = []string{}
			}
			dependents[dep] = append(dependents[dep], v)
		}
	}

	return &Scheduler{
		dependents: dependents,
		inDegree:   inDeg,
		total:      len(inDeg),
	}
}

// Run executes the scheduled tasks using up to workerCount concurrent
// workers. The exec callback is invoked for each task. If exec returns an
// error the scheduler cancels remaining work and returns that error. The
// returned slice contains task names in the order they were completed.
func (s *Scheduler) Run(ctx context.Context, nWorkers int, exec func(string) error) ([]string, error) {
	if nWorkers <= 0 {
		nWorkers = 1
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	taskCh := make(chan string)
	doneCh := make(chan string)
	errCh := make(chan error, 1)

	var wg sync.WaitGroup

	// worker pool
	for range nWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case t, ok := <-taskCh:
					if !ok {
						return
					}
					if err := exec(t); err != nil {
						select {
						case errCh <- err:
						default:
						}
						cancel()
						return
					}
					select {
					case doneCh <- t:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	var mutex sync.Mutex
	completed := []string{}

	// find tasks that can be immediately enqueued
	ready := make([]string, 0)
	for n, deg := range s.inDegree {
		if deg == 0 {
			ready = append(ready, n)
		}
	}

	// feed tasks into the workers
	go func() {
		defer close(taskCh)
		// enqueue initial ready list
		mutex.Lock()
		for _, t := range ready {
			select {
			case taskCh <- t:
			case <-ctx.Done():
				mutex.Unlock()
				return
			}
		}
		// clear ready so it's not enqueued again
		ready = nil
		mutex.Unlock()

		// now wait for completions and enqueue newly-ready tasks
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-doneCh:
				if !ok {
					return
				}
				mutex.Lock()
				// mark completed
				completed = append(completed, d)
				// decrement dependents
				for _, dep := range s.dependents[d] {
					s.inDegree[dep]--
					if s.inDegree[dep] == 0 {
						select {
						case taskCh <- dep:
						case <-ctx.Done():
							mutex.Unlock()
							return
						}
					}
				}
				mutex.Unlock()

				// if we've completed all nodes, we're done
				if len(completed) >= s.total {
					return
				}
			}
		}
	}()

	// wait for workers to finish (they will exit when taskCh is closed)
	wg.Wait()

	// check for error
	select {
	case err := <-errCh:
		return completed, err
	default:
	}

	// if not all tasks completed, there may be a cycle. If this ever happens,
	// we've a bug as the topological sort we do at the beginning ought to
	// find these.
	if len(completed) < s.total {
		return completed, fmt.Errorf("not all tasks completed: %d/%d", len(completed), s.total)
	}

	return completed, nil
}
