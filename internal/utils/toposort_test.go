package utils

import (
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

	result := TopologicalSort(vertices)

	// As maps in Go are a little probabalistic in their ordering, even with
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
