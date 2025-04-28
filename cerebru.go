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
	// EnableCleaner indicates whether the periodic cache cleaner is enabled.
	// The cleaner is responsible for removing expired nodes to free up
	// resources and maintain optimal performance.
	EnableCleaner bool

	// EnableDynamicSharding indicates whether dynamic sharding is enabled.
	// Dynamic sharding automatically adjusts the number of shards
	// based on usage patterns to optimize performance and resource
	// utilization.
	EnableDynamicSharding bool

	// ShardCap specifies the maximum number of shards allowed.
	// When dynamic sharding is enabled, the number of shards can
	// adjust based on usage patterns, but it will not exceed this limit.
	// When dynamic sharding is disabled, a fixed number of shards
	// will be created at initialization, each with the specified NodeCap.
	ShardCap int

	// NodeCap specifies the capacity of each node within a shard.
	// This value is fixed and does not change, regardless of whether
	// dynamic sharding is enabled or disabled.
	NodeCap int

	// MaxCost is an experimental field that defines the maximum cost
	// allowed for an operation or resource allocation, measured in
	// arbitrary units. It helps manage resource consumption and should
	// be set to a non-negative value. Use with caution, as this feature
	// is experimental and may change in future versions.
	// default:512
	MaxCost uint64
}

// New creates a new instance of CacheManager based on the provided configuration options.
// It initializes the cache manager with a specified number of shards and sets up the
// jump consistent hashing (JCH) for shard management.
func New(opt *Config) *CacheManager {
	var initialShards int
	var defaultMaxCost uint64

	if opt.MaxCost == 0 {
		defaultMaxCost = 512 * UnitMB
	} else {
		defaultMaxCost = opt.MaxCost
	}

	if opt.EnableDynamicSharding {
		initialShards = 4
	} else {
		initialShards = opt.ShardCap
	}

	manager := &CacheManager{
		pool:                      make([]*NodeShards, 0, opt.ShardCap),
		enableAutoCleaner:         opt.EnableCleaner,
		enableDynamicShardScaling: opt.EnableDynamicSharding,
		shardCap:                  opt.ShardCap,
		nodeCap:                   opt.NodeCap,
		jch:                       crypt.NewjCH(opt.ShardCap),
		maxCost:                   defaultMaxCost,
	}

	for i := 0; i < initialShards; i++ {
		manager.addShard()
	}

	return manager
}

// Set adds a key-value pair to the cache. If the key already exists, it updates the value.
// If the cache is full, it finds the least loaded shard to store the new entry.
// The entry is set to expire after 12 hours by default if it is newly created.
func (m *CacheManager) Set(key string, val interface{}, size uint64) {
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
		if node.expiredAt >= 0 && node.expiredAt <= time.Now().Unix() {
			shard.shardSize -= size
			shard.removeNode(node)
			delete(shard.pool, key)
			shard.size--
		} else {
			node.Value = val
			node.nodeSize = size
			node.expiredAt = time.Now().Add(12 * time.Hour).Unix()
			shard.mut.Unlock()
			return
		}
	}

	newNode := &Nodes{
		Key:      key,
		Value:    val,
		nodeSize: size,
	}

	shard.shardSize += size
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

// SetTTL adds a key-value pair to the cache with a specified time-to-live (TTL).
// If the key already exists, it updates the value and the expiration time.
// If the cache is full, it finds the least loaded shard to store the new entry.
func (m *CacheManager) SetTTL(key string, val interface{}, size uint64, ttl time.Duration) {
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
		shard.shardSize -= node.nodeSize
		node.Value = val
		node.nodeSize = size
		shard.shardSize += size
		node.expiredAt = expiry
		shard.moveToHead(node)
		shard.mut.Unlock()
		return
	}

	newNode := &Nodes{
		Key:       key,
		Value:     val,
		expiredAt: expiry,
		nodeSize:  size,
	}
	shard.shardSize += size
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

// Get retrieves the value associated with the given key from the cache.
// If the key exists and has not expired, it returns the value; otherwise, it returns nil.
func (m *CacheManager) Get(key string) interface{} {
	hashVal := m.jch.Hash(key)
	shardIndex := hashVal % uint64(len(m.pool))
	shard := m.pool[shardIndex]

	shard.mut.Lock()

	node, exists := shard.pool[key]
	if exists {
		now := time.Now().Unix()
		if node.expiredAt > 0 && node.expiredAt <= now {
			shard.shardSize -= node.nodeSize
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

// Remove deletes the key-value pair associated with the given key from the cache.
// It also removes the node from the eviction heap if it exists.
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
