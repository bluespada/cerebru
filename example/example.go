// Copyright (c) 2025 Bluespada <pentingmain@gmail.com>
//
// Distribute under MIT License, please read accompanying
// file copy or read online at https://opensource.org/license/mit

package main

import (
	"fmt"
	"time"

	"github.com/bluespada/cerebru"
)

func main() {
	// Initialize memory with configuration
	mem := cerebru.New(&cerebru.Config{
		// It is recommended to enable the cleaner to help
		// clean up expired cache entries if you use TTL.
		EnableCleaner: true,
		// Enable this option if you are dealing with unpredictable
		// cache entries and do not care about performance, but need
		// stability with a higher cache hit rate.
		EnableDynamicSharding: false,
		// ShardCap defines the maximum number of shards that can be created.
		// Each shard can help distribute the load and improve performance.
		ShardCap: 8,
		// NodeCap defines the maximum number of nodes that can be created
		// within each shard. This helps in managing the data distribution
		// and ensures that the system can scale effectively.
		NodeCap: 8,
	})

	// Set TTL for the key "key:totp"
	// Remember that Cerebru's eviction policy is quite strict,
	// so this key may be removed before it expires if there are
	// abnormal write operations.
	mem.SetTTL(
		"key:totp",
		"some-otp-ke",
		uint64(len([]byte("some-otp-ke"))),
		5*time.Second,
	)

	// Set data for the key "key:messages"
	mem.Set(
		"key:messages",
		&struct {
			Name string
			Age  int
		}{
			Name: "Yanto",
			Age:  12,
		},
		0, // If set to zero, when dynamic sharding is enabled, this key will be excluded from counting.
	)

	// Retrieve the value from the key "key:totp"
	res := mem.Get("key:totp")

	// Check if the result is not nil and print it
	if res != nil {
		fmt.Println(res)
	}
}
