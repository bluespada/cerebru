# Cerebru

**Cerebru** is an experimental in-memory caching library for Go, initially developed for internal projects.  
It focuses on high-performance cache management using **touch-based TTL** eviction and **sharded linked lists** to optimize memory usage and entry eviction.

> **Note:**  
> This project is under active development and **is not yet production-ready**. Use at your own risk.

## Features

- **Sharded Cache Architecture**  
  Data is partitioned across multiple shards to improve concurrency and reduce lock contention.

- **Touch-based TTL Eviction**  
  Entries are managed based on their last access time, with automatic expiration after a configurable TTL.

- **Optional Background Cleaner**  
  An optional background process that periodically removes expired entries with minimal performance impact.

- **Minimal and Focused API**  
  A straightforward API designed for general caching needs without unnecessary complexity.

- **Configurable Behavior**  
  Fully customizable settings for shard count, entry capacity, TTLs, and cleaning strategies.

- **Open Source (MIT License)**  
  Freely available under the permissive MIT license.

## Quick Benchmark

### Test Scenario

The benchmark simulates a high-concurrency cache workload:

- **Requests**: 500 million cache operations
- **Concurrency**: 2,000 concurrent workers
- **Pattern**: A mixed-access pattern with periodic updates every 1,000 operations.

The goal is to evaluate:
- **Cache efficiency** (hit and miss rates)
- **Memory consumption**
- **Garbage Collection overhead**

### Results

| Mode                     | Hit Rate  | Cache Hits      | Cache Misses     | Miss Rate        | Memory Allocated | System Memory | GC Cycles |
|---------------------------|-----------|-----------------|------------------|------------------|------------------|---------------|-----------|
| **Dynamic Sharding OFF**  | 95.97%    | 479,863,800     | 15,222,231       | 3.04%            | 35.96 MB         | 119.99 MB     | 1         |
| **Dynamic Sharding ON**   | 99.25%    | 496,272,177     | 263,013          | 0.05%            | 162.59 MB        | 821.03 MB     | 2         |

> **Interpretation:**  
> - Enabling dynamic sharding improves the hit rate significantly but increases memory usage and GC cycles, making it suitable for workloads with fluctuating traffic.  
> - Disabling dynamic sharding minimizes memory overhead and GC frequency, but results in a slightly lower hit rate, better suited for stable, predictable access patterns.

## Project Status

Cerebru is currently:
- Stable for internal testing
- Not yet validated in production environments
- Subject to API and behavior changes as development progresses

Contributions, feedback, and experimentation are welcome.

## License

This project is licensed under the [MIT License](LICENSE).