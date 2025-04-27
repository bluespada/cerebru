// Copyright (c) 2025 Bluespada <pentingmain@gmail.com>
//
// Distribute under MIT License, please read accompanying
// file copy or read online at https://opensource.org/license/mit

package cerebru

// Nodes represents a single entry in the cache.
// Each node contains a key-value pair, pointers for linked list traversal,
// and metadata for cache management policies.
type Nodes struct {
	// Key is the unique identifier for the cache entry.
	Key string

	// Value holds the data associated with the cache entry.
	Value interface{}

	// prev is a pointer to the previous node in the linked list,
	// allowing for bidirectional traversal of the cache entries.
	prev *Nodes

	// next is a pointer to the next node in the linked list,
	// allowing for bidirectional traversal of the cache entries.
	next *Nodes

	// expiredAt is the timestamp (in Unix time) indicating when the cache entry
	// should expire and be considered invalid.
	expiredAt int64

	// lastUsed is the timestamp (in Unix time) of the last time the cache entry
	// was accessed or modified. This is useful for eviction policies.
	lastUsed int64

	// sizeOfvalue represents the size of the value stored in this node,
	// which can be useful for managing memory and cache size limits.
	sizeOfvalue uint64
}
