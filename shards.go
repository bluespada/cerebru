// Copyright (c) 2025 Bluespada <pentingmain@gmail.com>
//
// Distribute under MIT License, please read accompanying
// file copy or read online at https://opensource.org/license/mit

package cerebru

import (
	"container/heap"
	"sync"
	"time"
)

// NodeShards represents a shard in a sharded cache system.
// It contains a pool of cache nodes, synchronization mechanisms, and metadata
// for managing eviction and node indexing.
type NodeShards struct {
	// pool holds the slice of Nodes that are part of this shard.
	pool map[string]*Nodes

	// mut is a read-write mutex used to synchronize access to the shard.
	// It allows multiple readers or a single writer to access the shard concurrently.
	mut sync.RWMutex

	// head is a pointer to the first node in the linked list of nodes within this shard.
	head *Nodes

	// tail is a pointer to the last node in the linked list of nodes within this shard.
	tail *Nodes

	capacity, size int

	// evictionHeap is a min-heap that manages the eviction of nodes based on their
	// last used timestamps, facilitating LRU or LFU eviction policies.
	evictionHeap *EvictionHeap

	// cleanerStop is a channel used to signal the stopping of background cleaning processes
	// that may be running to remove expired or unused nodes from the shard.
	cleanerStop chan struct{}

	// nodeIndex is a map that associates each node with its index in the pool,
	// allowing for efficient lookups and removals.
	nodeIndex map[*Nodes]int

	shardSize uint64
}

// addToHead adds a node to the head of the linked list in the NodeShards.
// It updates the node's previous and next pointers, sets the last used timestamp,
// and pushes the node onto the eviction heap.
func (ns *NodeShards) addToHead(node *Nodes) {
	now := time.Now().Unix()
	node.prev = ns.head
	nextNode := ns.head.next
	node.next = nextNode
	nextNode.prev = node
	ns.head.next = node
	node.lastUsed = now
	heap.Push(ns.evictionHeap, node)
	ns.nodeIndex[node] = ns.evictionHeap.Len() - 1
}

// moveToHead moves a node to the head of the linked list.
// It first removes the node from its current position and then adds it to the head.
func (ns *NodeShards) moveToHead(node *Nodes) {
	ns.removeNode(node)
	ns.addToHead(node)
}

// removeNode removes a node from the linked list and the eviction heap.
// It updates the node index and decreases the size of the NodeShards.
func (ns *NodeShards) removeNode(node *Nodes) {
	ns.removeFromTail(node)
	heap.Remove(ns.evictionHeap, ns.nodeIndex[node])
	delete(ns.nodeIndex, node)
}

// removeTail removes the last node from the linked list and returns it.
// It calls removeNode to handle the removal process.
func (ns *NodeShards) removeTail() *Nodes {
	node := ns.tail.prev
	ns.removeNode(node)
	return node
}

// startCleaner starts a background cleaner that periodically checks for expired nodes.
// It adjusts the cleaning interval based on the number of expired nodes found.
func (s *NodeShards) startCleaner() {
	baseInterval := time.Second * 5
	interval := baseInterval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			expiredCount := s.cleanExpired()
			if expiredCount == 0 {
				interval *= 2
				if interval > time.Minute {
					interval = time.Minute
				}
			} else {
				interval = baseInterval
			}
			ticker.Reset(interval)
		case <-s.cleanerStop:
			return
		}
	}
}

// cleanExpired checks for expired nodes in the pool and removes them.
// It also evicts nodes from the eviction heap if the size exceeds the capacity.
// Returns the count of expired nodes removed.
func (ns *NodeShards) cleanExpired() int {
	now := time.Now().Unix()
	expiredCount := 0

	ns.mut.Lock()
	defer ns.mut.Unlock()
	for key, node := range ns.pool {
		if node.expiredAt > 0 && node.expiredAt <= now {
			ns.removeNode(node)
			delete(ns.pool, key)
			expiredCount++
		}
	}

	if ns.evictionHeap == nil {
		return expiredCount
	}

	for ns.size > ns.capacity {
		evictedNode := heap.Pop(ns.evictionHeap)
		if evictedNode != nil {
			delete(ns.pool, evictedNode.(*Nodes).Key)
		}
	}
	return expiredCount
}

// moveToTail moves the last node in the linked list to the tail.
// It removes the node from the tail and updates the eviction heap.
func (shard *NodeShards) moveToTail() {
	if shard.tail.prev != shard.head {
		nodeToEvict := shard.tail.prev
		shard.removeFromTail(nodeToEvict)
		shard.evictionHeap.Pop()
	}
}

// removeFromTail removes a node from the tail of the linked list.
// It updates the pointers of the surrounding nodes to maintain the linked list structure.
func (shard *NodeShards) removeFromTail(node *Nodes) {
	if node.prev == nil {
		return
	}

	if node.prev.next != nil && shard.tail != nil {
		node.prev.next = shard.tail
	}

	if shard.tail.prev != nil && node.prev != nil {
		shard.tail.prev = node.prev
	}
}
