// Copyright (c) 2025 Bluespada <pentingmain@gmail.com>
//
// Distribute under MIT License, please read accompanying
// file copy or read online at https://opensource.org/license/mit
package cerebru

import "container/heap"

// EvictionHeap is a min-heap of Nodes pointers, which allows for efficient
// retrieval and removal of the least recently used nodes.
type EvictionHeap []*Nodes

// Len returns the number of nodes in the eviction heap.
func (eh EvictionHeap) Len() int {
	return len(eh)
}

// Less reports whether the node at index next should sort before the node at index prev.
// It compares the lastUsed timestamps of the nodes.
func (eh EvictionHeap) Less(next, prev int) bool {
	if next >= eh.Len() || prev >= eh.Len() {
		return false
	}

	if eh[next] == nil && eh[prev] == nil {
		return false
	} else if eh[next] == nil {
		return false
	} else if eh[prev] == nil {
		return true
	}

	return eh[next].lastUsed < eh[prev].lastUsed
}

// Pop removes and returns the node with the least recently used timestamp
// from the eviction heap.
func (eh *EvictionHeap) Pop() interface{} {
	old := *eh
	new := len(old)
	if new == 0 {
		return nil
	}
	item := old[new-1]
	*eh = old[:new-1]
	return item
}

// Swap exchanges the nodes at indices next and prev in the eviction heap.
func (eh EvictionHeap) Swap(next, prev int) {
	if next < 0 || next >= len(eh) || prev < 0 || prev >= len(eh) {
		return
	}
	eh[next], eh[prev] = eh[prev], eh[next]
}

// Push adds a new node to the eviction heap.
func (eh *EvictionHeap) Push(node interface{}) {
	*eh = append(*eh, node.(*Nodes))
}

// RemoveNode removes a specific node from the eviction heap.
// It reinitializes the heap to maintain the heap property after removal.
func (eh *EvictionHeap) RemoveNode(node *Nodes) {
	for i, n := range *eh {
		if n == node {
			*eh = append((*eh)[:i], (*eh)[i+1:]...)
			heap.Init(eh)
			return
		}
	}
}
