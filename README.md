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

### Test Case

The provided test simulates a high-concurrency cache load, similar to what platforms like Google or Netflix might experience. Here’s a simplified explanation:

- **Goal**: Mimic high traffic by simulating **500 million requests** with **2,000 concurrent workers**. This tests how well the cache handles real-world, high-volume requests.
  
- **How it works**: 
  - **Cache Hits**: When a request for a key finds it in the cache.
  - **Cache Misses**: When the key isn’t in the cache and gets added.
  - Every 1,000th request sets a special cache value, simulating periodic updates.

- **Simulation Context**: The test models real-world traffic load seen by large-scale services, testing **cache efficiency** and **concurrency management** under heavy use.

This kind of test is useful for validating how a caching system will perform under the stress of millions of simultaneous requests, much like what you’d find in production environments at major tech companies.

Below are the benchmark results that show the difference in memory usage, **GC cycles**, and **hit rate** between the two scenarios:


| Mode                     | Hit Rate  | Cache Hit Count | Cache Miss Count | Cache Miss Rate | Memory Allocated | Total Heap Allocated | System Memory | Num GC Cycles |
|--------------------------|-----------|-----------------|------------------|-----------------|------------------|----------------------|----------------|---------------|
| **Dynamic Sharding OFF**  | 95.97%    | 479,863,800     | 15,222,231       | 3.04%           | 35.96 MB         | 35.96 MB            | 119.99 MB      | 1             |
| **Dynamic Sharding ON**   | 99.25%    | 496,272,177     | 263,013          | 0.05%           | 162.59 MB        | 162.59 MB           | 821.03 MB      | 2             |


> **Note:**  
> - **Dynamic Sharding ON** improves **hit rate** but increases **memory usage** and **GC cycles**. This is beneficial for applications with **unpredictable data** or **fluctuating traffic patterns** that require stable distribution and quick adaptation.  
> - **Dynamic Sharding OFF** reduces **memory overhead** and **GC cycles**, but results in a **lower hit rate** and **higher miss rate**. It’s better suited for applications with **predictable access patterns** or **steady workloads**.
> - The performance trade-offs will depend heavily on the application's **data volatility** and **request frequency**.


# License
This project is licensed under the [MIT License](LICENSE).