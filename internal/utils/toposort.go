package utils

import "errors"

var ErrCycleDetected = errors.New("detected cycle in DAG")

// TopologicalSort takes a map representing a DAG and linearises the DAG based
// the connected vertices (given in the values of the associated list).
//
// The resulting list can be consumed from the end as the last element will
// have no other dependencies.
//
// Taken (with naming changes) from
// https://tylercipriani.com/blog/2017/09/13/topographical-sorting-in-golang/
func TopologicalSort[K comparable](graph map[K][]K) ([]K, error) {
	result := []K{}

	// inDegress keeps track of the number of incoming edges
	inDegree := map[K]int{}
	for n := range graph {
		inDegree[n] = 0
	}
	// count the incoming edges for each node
	for _, adjacent := range graph {
		for _, v := range adjacent {
			inDegree[v]++
		}
	}

	next := []K{}
	for u, v := range inDegree {
		if v != 0 {
			continue
		}

		next = append(next, u)
	}

	for len(next) > 0 {
		u := next[0]
		next = next[1:]

		result = append(result, u)

		for _, v := range graph[u] {
			inDegree[v]--
			if inDegree[v] == 0 {
				next = append(next, v)
			}
		}
	}

	// We can detect if there are fewer elements in the result than in the graph
	if len(result) < len(graph) {
		return []K{}, ErrCycleDetected
	}

	return result, nil
}
