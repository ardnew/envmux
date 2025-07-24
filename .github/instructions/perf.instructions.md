---
description: 'Instructions for writing code in any language that optimizes runtime performance'
applyTo: '*'
---

# Performance Optimization Best Practices

---

## General Principles

- **Measure First, Optimize Second:** Always profile and measure before optimizing. Use benchmarks, profilers, and monitoring tools to identify real bottlenecks. Guessing is the enemy of performance.
  - *Pro Tip:* Prefer tools from your language's standard toolchain. Use the built-in pprof tool to profile Go code.
- **Optimize for the Common Case:** Focus on optimizing code paths that are most frequently executed. Don't waste time on rare edge cases unless they're critical.
- **Avoid Premature Optimization:** Write clear, maintainable code first; optimize only when necessary. Premature optimization can make code harder to read and maintain.
- **Minimize Resource Usage:** Use memory, CPU, network, and disk resources efficiently. Always ask: "Can this be done with less?"
- **Prefer Simplicity:** Simple algorithms and data structures are often faster and easier to optimize. Don't over-engineer.
- **Document Performance Assumptions:** Clearly comment on any code that is performance-critical or has non-obvious optimizations. Future maintainers (including you) will thank you.
- **Understand the Platform:** Know the performance characteristics of your language, framework, and runtime. What's fast in Python may be slow in JavaScript, and vice versa.
- **Automate Performance Testing:** Integrate performance tests and benchmarks into your CI/CD pipeline. Catch regressions early.
- **Set Performance Budgets:** Define acceptable limits for load time, memory usage, API latency, etc. Enforce them with automated checks.

---

## Backend Performance

### Algorithm and Data Structure Optimization
- **Choose the Right Data Structure:** Arrays for sequential access, maps for fast lookups, trees for hierarchical data, etc.
- **Efficient Algorithms:** Use binary search, quicksort, or hash-based algorithms where appropriate.
- **Avoid O(n^2) or Worse:** Profile nested loops and recursive calls. Refactor to reduce complexity.
- **Batch Processing:** Process data in batches to reduce overhead (e.g., bulk database inserts).
- **Streaming:** Use streaming techniques for communicating data concurrently.
- **Resource Cleanup:** Always release resources (files, sockets, DB connections) promptly.

### Concurrency and Parallelism
- **Prefer Concurrency Primitives:** If applicable, use native concurrency resources such as channels and pipes instead artificial or manual synchronization that depends on execution timing, memory barriers, critical sections, etc.
- **Share Memory By Communicating:** Do not communicate by sharing memory; instead, share memory by communicating.
- **Backpressure:** Implement backpressure in queues and pipelines to avoid overload.

### Logging and Monitoring
- **Minimize Logging in Hot Paths:** Excessive logging can slow down critical code.
- **Structured Logging:** Use JSON or key-value logs for easier parsing and analysis.
- **Monitor Everything:** Latency, throughput, error rates, resource usage.
- **Alerting:** Set up alerts for performance regressions and resource exhaustion.

---

## Code Review Checklist for Performance

- [ ] Are there any obvious algorithmic inefficiencies (O(n^2) or worse)?
- [ ] Are data structures appropriate for their use?
- [ ] Are there unnecessary computations or repeated work?
- [ ] Is caching used where appropriate, and is invalidation handled correctly?
- [ ] Are there any memory leaks or unbounded resource usage?
- [ ] Are there any blocking operations in hot paths?
- [ ] Is logging in hot paths minimized and structured?
- [ ] Are performance-critical code paths documented and tested?
- [ ] Are there automated tests or benchmarks for performance-sensitive code?
- [ ] Are there alerts for performance regressions?

---

## Advanced Topics

### Memory Management
- **Heap Monitoring:** Monitor heap usage and garbage collection.

### Scalability
- **Bottleneck Analysis:** Identify and address single points of failure.
- **Distributed Systems:** Use idempotent operations, retries, and circuit breakers.

### Security and Performance
- **Efficient Crypto:** Use hardware-accelerated and well-maintained cryptographic libraries.
- **Validation:** Validate inputs efficiently; avoid regexes in hot paths.

---

## References and Further Reading
- [How To Write Go Code](https://go.dev/doc/code)
- [Effective Go](https://go.dev/doc/effective_go)
