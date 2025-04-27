// Copyright (c) 2025 Bluespada <pentingmain@gmail.com>
//
// Distribute under MIT License, please read accompanying
// file copy or read online at https://opensource.org/license/mit

package cerebru

import (
	"container/heap"
	"time"

	"github.com/bluespada/cerebru/internal/crypt"
)

// Config holds the configuration options for the CacheManager.
// It includes flags for enabling the cleaner and dynamic sharding,
// as well as capacity settings for shards and nodes.
type Config struct {
	EnableCleaner, EnableDynamicSharding bool // Flags to enable the cache cleaner and dynamic sharding
	ShardCap, NodeCap                    int  // Capacity settings for shards and nodes
}

// New creates a new instance of CacheManager based on the provided configuration options.
// It initializes the cache manager with a specified number of shards and sets up the
// jump consistent hashing (JCH) for shard management.
func New(opt *Config) *CacheManager {
	var initialShards int

	if opt.EnableDynamicSharding {
		initialShards = 4
	} else {
		initialShards = opt.ShardCap
	}

	manager := &CacheManager{
		pool:                      make([]*NodeShards, 0, initialShards),
		enableAutoCleaner:         opt.EnableCleaner,
		enableDynamicShardScaling: opt.EnableDynamicSharding,
		shardCap:                  opt.ShardCap,
		nodeCap:                   opt.NodeCap,
		jch:                       crypt.NewjCH(opt.ShardCap),
	}

	for i := 0; i < initialShards; i++ {
		manager.addShard()
	}

	return manager
}

func (m *CacheManager) Set(key string, val interface{}) {

	if m.enableDynamicShardScaling {
		m.dynamicShardScaling()
	}

	hashVal := m.jch.Hash(key)
	shardIndex := hashVal % uint64(len(m.pool))
	shard := m.pool[shardIndex]

	shard.mut.Lock()
	if shard.size >= shard.capacity {
		shard.mut.Unlock()
		shard = m.findLeastLoadedShard()
		shard.mut.Lock()
	}

	if node, exists := shard.pool[key]; exists {
		if node.expiredAt > 0 && node.expiredAt <= time.Now().Unix() {
			shard.removeNode(node)
			delete(shard.pool, key)
			shard.size--
		} else {
			node.Value = val
			node.expiredAt = time.Now().Add(12 * time.Hour).Unix()
			shard.mut.Unlock()
			return
		}
	}

	newNode := &Nodes{
		Key:   key,
		Value: val,
	}

	shard.addToHead(newNode)
	shard.pool[key] = newNode
	shard.size++

	if shard.size > shard.capacity {
		evicted := shard.removeTail()
		if evicted != nil {
			delete(shard.pool, evicted.Key)
			shard.size--
		}
	}

	shard.mut.Unlock()
}

func (m *CacheManager) SetTTL(key string, val interface{}, ttl time.Duration) {
	if m.enableDynamicShardScaling {
		m.dynamicShardScaling()
	}

	hashVal := m.jch.Hash(key)
	shardIndex := hashVal % uint64(len(m.pool))
	shard := m.pool[shardIndex]

	shard.mut.Lock()
	if shard.size >= shard.capacity {
		shard.mut.Unlock()
		shard = m.findLeastLoadedShard()
		shard.mut.Lock()
	}

	expiry := int64(0)
	if ttl > 0 {
		expiry = time.Now().Add(ttl).Unix()
	}

	if node, exists := shard.pool[key]; exists {
		node.Value = val
		node.expiredAt = expiry
		shard.moveToHead(node)
		shard.mut.Unlock()
		return
	}

	newNode := &Nodes{
		Key:       key,
		Value:     val,
		expiredAt: expiry,
	}

	shard.addToHead(newNode)
	shard.pool[key] = newNode
	shard.size++

	if shard.size > shard.capacity {
		evicted := shard.removeTail()
		if evicted != nil {
			delete(shard.pool, evicted.Key)
			shard.size--
		}
	}

	shard.mut.Unlock()
}

func (m *CacheManager) Get(key string) interface{} {
	hashVal := m.jch.Hash(key)
	shardIndex := hashVal % uint64(len(m.pool))
	shard := m.pool[shardIndex]

	shard.mut.Lock()

	node, exists := shard.pool[key]
	if exists {
		now := time.Now().Unix()
		if node.expiredAt > 0 && node.expiredAt <= now {
			shard.removeNode(node)
			delete(shard.pool, key)
			shard.size--
			shard.mut.Unlock()
			return nil
		}
		shard.moveToHead(node)
		shard.mut.Unlock()
		return node.Value
	}
	shard.mut.Unlock()
	return nil
}

func (m *CacheManager) Remove(key string) {

	hashVal := m.jch.GetBucket(key)
	shardIndex := hashVal % uint64(len(m.pool))
	shard := m.pool[shardIndex]

	shard.mut.Lock()
	defer shard.mut.Unlock()

	if node, exists := shard.pool[key]; exists {
		shard.evictionHeap.RemoveNode(node)
		shard.removeNode(node)
		delete(shard.pool, key)
		shard.size--
		heap.Init(shard.evictionHeap)
	}
}
