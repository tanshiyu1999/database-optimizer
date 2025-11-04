# CSV Import Optimization: Sequential vs Concurrent

## Overview

This module demonstrates two approaches to importing large CSV files into PostgreSQL:

1. **Sequential Import** (`ImportFromFile`) - Single-threaded, simpler
2. **Concurrent Import** (`ImportFromFileGoRoutine`) - Multi-threaded, faster

## Sequential Import (Original)

### How It Works
```
CSV File → Read Row → Parse Row → Add to Batch → Insert Batch → Repeat
```

**Characteristics:**
- ✅ Simple and predictable
- ✅ Easy to debug
- ✅ Maintains exact insertion order
- ❌ CPU-bound (single core usage)
- ❌ Slower for large datasets

**Best For:**
- Small to medium datasets (< 100K rows)
- When debugging issues
- When order matters strictly
- Learning database basics

### Usage
```go
importer := importer.NewCSVImporter(db.GetPool(), 1000)
count, err := importer.ImportFromFile("data/sf-fire-calls.csv")
```

## Concurrent Import (Optimized)

### How It Works
```
                    ┌─→ Worker 1 (Parse) ─┐
CSV File → Reader ──┼─→ Worker 2 (Parse) ─┼─→ Insert Pool → Database
                    └─→ Worker N (Parse) ─┘
```

**Pipeline Stages:**
1. **Stage 1**: Single reader goroutine reads CSV sequentially
2. **Stage 2**: N worker goroutines parse records into batches
3. **Stage 3**: M insert goroutines write batches to database concurrently

**Characteristics:**
- ✅ Much faster (2-4x speed improvement typical)
- ✅ Better CPU utilization (multi-core)
- ✅ Non-blocking pipeline architecture
- ❌ More complex to debug
- ❌ Slightly more memory usage
- ❌ Order not guaranteed (but doesn't matter for bulk imports)

**Best For:**
- Large datasets (100K+ rows)
- Production imports
- Time-sensitive operations
- Learning concurrent programming

### Usage
```go
importer := importer.NewCSVImporter(db.GetPool(), 1000)
numWorkers := 4 // Adjust based on CPU cores
count, err := importer.ImportFromFileGoRoutine("data/sf-fire-calls.csv", numWorkers)
```

## Performance Comparison

### Test Setup
- Dataset: 175,298 SF Fire Call records
- Batch Size: 1000 records
- Hardware: Modern multi-core CPU
- Database: PostgreSQL 16 (Docker)

### Expected Results

| Method | Time | Records/sec | CPU Usage |
|--------|------|-------------|-----------|
| Sequential | ~15-20s | 8,000-12,000 | 1 core (~25%) |
| Concurrent (2 workers) | ~8-12s | 14,000-22,000 | 2 cores (~50%) |
| Concurrent (4 workers) | ~6-10s | 17,000-29,000 | 4 cores (~100%) |
| Concurrent (8 workers) | ~6-9s | 19,000-29,000 | Diminishing returns |

**Note:** Actual performance depends on:
- CPU cores available
- Disk I/O speed
- Network latency to database
- PostgreSQL configuration
- Presence of indexes

## Student Exercise: Benchmark Both Methods

### Task 1: Compare Performance

1. **Run Sequential Import**
   ```go
   // In main.go, use:
   recordsImported, err := csvImporter.ImportFromFile("data/sf-fire-calls.csv")
   ```
   Record the results:
   - Import time: ______
   - Records/sec: ______

2. **Run Concurrent Import (4 workers)**
   ```go
   recordsImported, err := csvImporter.ImportFromFileGoRoutine("data/sf-fire-calls.csv", 4)
   ```
   Record the results:
   - Import time: ______
   - Records/sec: ______

3. **Calculate Improvement**
   ```
   Speed increase = (Concurrent time / Sequential time) × 100%
   ```

### Task 2: Find Optimal Worker Count

Test with different worker counts:
```go
// Try: 1, 2, 4, 8, 16 workers
csvImporter.ImportFromFileGoRoutine("data/sf-fire-calls.csv", numWorkers)
```

| Workers | Time | Records/sec | Notes |
|---------|------|-------------|-------|
| 1 | | | |
| 2 | | | |
| 4 | | | |
| 8 | | | |
| 16 | | | |

**Questions to Answer:**
- At what point do you see diminishing returns?
- Why doesn't performance keep improving linearly?
- What's the bottleneck? (CPU, I/O, Database?)

### Task 3: Impact of Indexes

1. Run concurrent import **without indexes** (comment them out in schema.sql)
2. Run concurrent import **with indexes**
3. Compare the times

| Scenario | Time | Records/sec |
|----------|------|-------------|
| No Indexes | | |
| With Indexes | | |

**Why the difference?**

## How the Concurrent Version Works

### Code Architecture

```go
// Channel Pipeline
recordChan   := make(chan []string, batchSize*2)      // CSV records
batchChan    := make(chan [][]interface{}, numWorkers) // Parsed batches
errorChan    := make(chan error, numWorkers)          // Error handling
doneChan     := make(chan bool)                       // Completion signal
```

### Stage 1: CSV Reader (1 goroutine)
```go
go func() {
    defer close(recordChan)
    for {
        record, err := reader.Read()
        if err == io.EOF { break }
        recordChan <- record  // Send to workers
    }
}()
```

**Why single-threaded?**
- CSV readers are not thread-safe
- File I/O is already sequential
- Parsing is the bottleneck, not reading

### Stage 2: Parse Workers (N goroutines)
```go
for i := 0; i < numWorkers; i++ {
    go func(workerID int) {
        batch := make([][]interface{}, 0, batchSize)
        for record := range recordChan {
            row := parseRow(record)      // CPU-intensive
            batch = append(batch, row)
            if len(batch) >= batchSize {
                batchChan <- batch       // Send to inserters
            }
        }
    }(i)
}
```

**Why multiple workers?**
- Parsing (string manipulation, type conversion) is CPU-bound
- Multiple cores can work in parallel
- Each worker processes different records

### Stage 3: Insert Workers (M goroutines)
```go
for i := 0; i < insertWorkers; i++ {
    go func(workerID int) {
        for batch := range batchChan {
            imp.executeBatch(ctx, batch)  // Database insert
        }
    }(i)
}
```

**Why concurrent inserts?**
- PostgreSQL can handle multiple connections
- Network latency can be hidden
- Batches can be processed in parallel

## Common Issues & Solutions

### Issue 1: Too Many Workers
**Symptom:** Performance degrades with 16+ workers
**Cause:** Context switching overhead, connection pool exhaustion
**Solution:** Keep workers ≤ CPU cores × 2

### Issue 2: Memory Usage Spikes
**Symptom:** High memory consumption
**Cause:** Channel buffers too large
**Solution:** Reduce channel buffer sizes or batch size

### Issue 3: Inconsistent Performance
**Symptom:** Wide variance in import times
**Cause:** Database vacuuming, other processes, cold start
**Solution:** Run multiple tests, warm up database

## Advanced Optimizations (For Interested Students)

### 1. PostgreSQL COPY Command
Replace INSERT with COPY for 5-10x speed:
```go
conn.CopyFrom(ctx, pgx.Identifier{"fire_calls"}, columns, rows)
```

### 2. Disable Indexes During Import
```sql
DROP INDEX idx_call_type;  -- Drop before import
-- Import data
CREATE INDEX idx_call_type ON fire_calls(call_type);  -- Recreate after
```

### 3. Adjust Batch Size Dynamically
```go
// Smaller batches for indexes, larger without
batchSize := 5000  // No indexes
batchSize := 500   // With indexes
```

### 4. Use Transactions
```go
tx, _ := pool.Begin(ctx)
// Multiple inserts
tx.Commit(ctx)
```

## Learning Outcomes

After completing this module, you should understand:

- ✅ When to use concurrent vs sequential processing
- ✅ How Go channels enable pipeline architectures
- ✅ Trade-offs between complexity and performance
- ✅ How to profile and benchmark code improvements
- ✅ Database optimization strategies (batching, indexes, COPY)
- ✅ Diminishing returns with parallelization

## Further Reading

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [PostgreSQL Bulk Loading](https://www.postgresql.org/docs/current/populate.html)
- [Channel Best Practices](https://go.dev/doc/effective_go#channels)

---

**Remember:** Always measure! The "faster" approach isn't always better if it adds complexity you don't need for your dataset size.
