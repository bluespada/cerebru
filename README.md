# Cerebru

An experimental in-memory caching library written in Go, originally developed for the internal projects. Cerebru focuses on high-performance cache management using touch-based TTL and dynamic sharding strategies to optimize memory usage and eviction.  
While primarily intended for internal systems, it's made available under the MIT License for anyone needing a fast, fine-grained caching solution.

>[!IMPORTANT] 
> This project is a work in progress. It is **not yet production-ready**, and documentation is currently incomplete. Use at your own discretion.
## Features

- Dynamic Sharding  
  Adjusts the number of shards based on usage patterns, allowing more flexible scaling under load.

- Efficient Lookup  
  Optimized for maintaining a stable hit rate under various access patterns.

- Optional Cleaner  
  Background cleanup process for removing expired entries, designed to minimize performance impact.

- Minimal API Surface  
  Straightforward API focused on simplicity and ease of use without unnecessary complexity.

- Configurable Behavior  
  Fully adjustable settings for shard capacity, node capacity, cleaner operation, and sharding strategy.

- Open Source (MIT License)  
  Available for public use under a permissive MIT license.



## Quick Benchmark

**Test Case:**  
50,000 keys (lightweight string payload)  
Hardware: Standard consumer laptop (no special tuning)

| Mode               | Hit Rate | Alloc (Active) | TotalAlloc | Sys Memory | Num GC Cycles |
|--------------------|----------|----------------|------------|------------|--------------|
| Dynamic Sharding OFF | 9.45%   | 7 MiB          | 7 MiB      | 15 MiB     | 2            |
| Dynamic Sharding ON  | 99.95%  | 648 MiB        | 1912 MiB   | 924 MiB    | 22           |

> Note: Higher memory usage is expected with Dynamic Sharding enabled as the system adapts to heavier loads.



# License
This project is licensed under the [MIT License](LICENSE).