package main

import (
	"fmt"
	"sync"
	"time"

	"runtime"

	"github.com/bluespada/cerebru"
)

func main() {
	mem := cerebru.New(&cerebru.Config{
		EnableCleaner:         true,
		EnableDynamicSharding: false,
		ShardCap:              8,
		NodeCap:               8,
		MaxCost:               500 * cerebru.UnitMB,
	})

	var wg sync.WaitGroup
	const totalRequests = 500_000_000
	concurrency := 2_000

	var cacheHits, cacheMisses int
	start := time.Now()

	apiRequest := func(i int) {
		defer wg.Done()

		key := fmt.Sprintf("key:%d", i)
		res := mem.Get(key)
		if res != nil {
			cacheHits++
		} else {
			cacheMisses++
			mem.Set(key, fmt.Sprintf("value-%d", i), 0)
		}

		if i%1000 == 0 {
			mem.Set(key, fmt.Sprintf("ttl-value-%d", i), 0)
		}
	}

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go apiRequest(i % concurrency)
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Total time for %d requests: %s\n", totalRequests, elapsed)

	hitRate := float64(cacheHits) / float64(totalRequests) * 100
	missRate := float64(cacheMisses) / float64(totalRequests) * 100

	fmt.Printf("Cache Hit Count: %d\n", cacheHits)
	fmt.Printf("Cache Miss Count: %d\n", cacheMisses)
	fmt.Printf("Cache Hit Rate: %.2f%%\n", hitRate)
	fmt.Printf("Cache Miss Rate: %.2f%%\n", missRate)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	fmt.Println("--- System Usage ---")
	fmt.Printf("Memory Allocated: %.2f MB\n", float64(memStats.Alloc)/1024/1024)
	fmt.Printf("System Memory: %.2f MB\n", float64(memStats.Sys)/1024/1024)
	fmt.Printf("Total Heap Allocated: %.2f MB\n", float64(memStats.HeapAlloc)/1024/1024)
	fmt.Printf("CPU Cores: %d\n", runtime.NumCPU())
}
