package utils

import (
	"errors"
	"testing"
)

func TestTopologicalSort(t *testing.T) {
	// Directed Acyclic Graph
	vertices := map[int][]int{
		1:  {4},
		2:  {3},
		3:  {4, 5},
		4:  {6},
		5:  {6},
		6:  {7, 11},
		7:  {8},
		8:  {14},
		9:  {10},
		10: {11},
		11: {12},
		13: {13},
		14: {},
	}

	result, err := TopologicalSort(vertices)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// As maps in Go are non-deterministic in their ordering, even with
	// as keys, we can't just compare the result to a static list. Instead, we
	// check that each of the vertices comes earlier than the connected vertices.
	for i, v := range result {
		for _, connected := range vertices[v] {
			found := false
			for j := i; j < len(result); j++ {
				if connected == result[j] {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("For %v, dependency %v was not found later in the list; result: %v", i, connected, result)
			}
		}
	}
}

func TestForGraphCycle(t *testing.T) {
	vertices := map[int][]int{
		1: {2, 3},
		2: {},
		3: {2, 1},
	}

	if result, err := TopologicalSort(vertices); err == nil {
		t.Errorf("expected error, got %v", result)
	} else if !errors.Is(err, ErrCycleDetected) {
		t.Errorf("expected ErrCycleDetected, got %v", err)
	}
}
