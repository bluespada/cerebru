// Copyright (c) 2025 Bluespada <pentingmain@gmail.com>
//
// Distribute under MIT License, please read accompanying
// file copy or read online at https://opensource.org/license/mit

package crypt

import (
	"hash/fnv"
)

// JCH represents a Jump Consistent Hashing structure.
// It manages the number of buckets and a cache for hashed keys.
type JCH struct {
	numBuckets int               // Number of buckets for hashing
	hashCache  map[string]uint64 // Cache to store computed hash values for keys
}

// NewjCH creates a new instance of JCH with the specified number of buckets.
func NewjCH(numBuckets int) *JCH {
	return &JCH{
		numBuckets: numBuckets,
		hashCache:  make(map[string]uint64),
	}
}

// Hash computes the hash value for a given key using FNV-1a hashing algorithm.
// It returns the 64-bit hash value.
func (j *JCH) Hash(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

// GetBucket retrieves the bucket index for a given key.
// It checks the hash cache first; if the key is not cached, it computes the hash,
// determines the bucket index, caches it, and returns the index.
func (j *JCH) GetBucket(key string) uint64 {
	if bucket, exists := j.hashCache[key]; exists {
		return bucket
	}

	h := j.Hash(key)
	bucket := h % uint64(j.numBuckets)
	j.hashCache[key] = bucket
	return bucket
}
