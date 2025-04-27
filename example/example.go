// Copyright (c) 2025 Bluespada <pentingmain@gmail.com>
//
// Distribute under MIT License, please read accompanying
// file copy or read online at https://opensource.org/license/mit

package main

import (
	"fmt"
	"runtime"
	"strconv"
	"sync/atomic"

	"github.com/bluespada/cerebru"
)

func main() {
	mem := cerebru.New(&cerebru.Config{
		EnableCleaner:         true,
		EnableDynamicSharding: true,
		ShardCap:              64,
		NodeCap:               64,
	})
	var (
		totalHit  uint64
		totalMiss uint64
	)

	for i := 0; i < 50_000; i++ {
		key := "key_2:" + strconv.Itoa(i)
		val := "yanto-number" + strconv.Itoa(i)
		mem.Set(key, val)
		if res := mem.Get(key); res != nil {
			atomic.AddUint64(&totalHit, 1)
		} else {
			atomic.AddUint64(&totalMiss, 1)
		}
	}

	// Final metrics after all tests
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	totalOps := totalHit + totalMiss
	hitRate := (float64(totalHit) / float64(totalOps)) * 100
	missRate := (float64(totalMiss) / float64(totalOps)) * 100

	fmt.Printf("\n--- Benchmark Summary ---\n")
	fmt.Printf("Total Ops: %d\n", totalOps)
	fmt.Printf("Cache Hit: %d\n", totalHit)
	fmt.Printf("Cache Miss: %d\n", totalMiss)
	fmt.Printf("Hit Rate: %.2f%%\n", hitRate)
	fmt.Printf("Miss Rate: %.2f%%\n", missRate)

	fmt.Printf("\n--- Memory Stats ---\n")
	fmt.Printf("Alloc = %v MiB\n", bToMb(memStats.Alloc))
	fmt.Printf("TotalAlloc = %v MiB\n", bToMb(memStats.TotalAlloc))
	fmt.Printf("Sys = %v MiB\n", bToMb(memStats.Sys))
	fmt.Printf("NumGC = %v\n", memStats.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
