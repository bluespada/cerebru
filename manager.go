// Copyright (c) 2025 Bluespada <pentingmain@gmail.com>
//
// Distribute under MIT License, please read accompanying
// file copy or read online at https://opensource.org/license/mit

package cerebru

import (
	"container/heap"
	"sync"

	"github.com/bluespada/cerebru/internal/crypt"
)

// CacheManager manages a pool of NodeShards for caching.
// It handles shard creation, dynamic scaling, and node rebalancing.
type CacheManager struct {
	pool                                         []*NodeShards
	enableAutoCleaner, enableDynamicShardScaling bool
	shardCap, nodeCap                            int
	poolMut                                      sync.RWMutex
	jch                                          *crypt.JCH
}

// addShard creates a new NodeShards instance and adds it to the pool.
// If auto-cleaning is enabled, it starts the cleaner for the new shard.
func (m *CacheManager) addShard() {
	m.poolMut.Lock()
	shard := &NodeShards{
		pool:         make(map[string]*Nodes),
		head:         &Nodes{},
		tail:         &Nodes{},
		capacity:     m.nodeCap,
		cleanerStop:  make(chan struct{}),
		evictionHeap: &EvictionHeap{},
		nodeIndex:    map[*Nodes]int{},
		mut:          sync.RWMutex{},
	}
	shard.head.next = shard.tail
	shard.tail.prev = shard.head
	heap.Init(shard.evictionHeap)

	if m.enableAutoCleaner {
		go shard.startCleaner()
	}

	m.pool = append(m.pool, shard)
	m.poolMut.Unlock()
}

// findLeastLoadedShard returns the shard with the least number of nodes.
func (m *CacheManager) findLeastLoadedShard() *NodeShards {
	var target *NodeShards
	minLoad := int(^uint(0) >> 1)

	for _, shard := range m.pool {
		count := len(shard.pool)
		if count < minLoad {
			minLoad = count
			target = shard
		}
	}
	return target
}

// dynamicShardScaling checks the load of shards and adds or removes shards as needed.
func (m *CacheManager) dynamicShardScaling() {
	var addShardNeeded bool
	var removeShardNeeded bool

	for _, shard := range m.pool {
		if shard.size >= m.nodeCap-1 {
			addShardNeeded = true
			break
		}
	}

	for _, shard := range m.pool {
		if shard.size <= m.nodeCap/4 && len(m.pool) > 2 {
			removeShardNeeded = true
			break
		}
	}

	if addShardNeeded {
		m.addShard()
		m.rebalanceNodes()
	}

	if removeShardNeeded {
		m.removeShardAndRebalance()
	}
}

// removeShardAndRebalance removes empty shards from the pool and rebalances nodes.
func (m *CacheManager) removeShardAndRebalance() {
	m.poolMut.Lock()
	defer m.poolMut.Unlock()

	for i := len(m.pool) - 1; i >= 0; i-- {
		shard := m.pool[i]
		if shard.size == 0 && len(m.pool) > 2 {
			m.pool = append(m.pool[:i], m.pool[i+1:]...)
		}
	}

	m.rebalanceNodes()
}

// rebalanceNodes redistributes nodes across shards to maintain balance.
func (m *CacheManager) rebalanceNodes() {
	totalNodes := 0
	for _, shard := range m.pool {
		shard.mut.Lock()
		totalNodes += len(shard.pool)
		shard.mut.Unlock()
	}

	allNodes := make([]*Nodes, 0, totalNodes)

	for _, shard := range m.pool {
		shard.mut.Lock()
		for _, node := range shard.pool {
			allNodes = append(allNodes, node)
		}

		for key := range shard.pool {
			delete(shard.pool, key)
		}
		shard.size = 0
		shard.mut.Unlock()
	}

	for _, node := range allNodes {
		idx := m.jch.Hash(node.Key) % uint64(len(m.pool))
		shard := m.pool[idx]

		shard.mut.Lock()

		if shard.size >= shard.capacity {
			shard.moveToTail()
		}

		shard.pool[node.Key] = node
		shard.addToHead(node)
		shard.size++

		shard.mut.Unlock()
	}
}
